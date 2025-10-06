package rgstat

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"os"

	pq "github.com/parquet-go/parquet-go"
)

type BasicRowGroupStat struct {
	NumRows        int64 `json:"number_of_rows"`
	ColumnChunks   int64 `json:"number_of_column_chunks"`
	SortingColumns int64 `json:"number_of_sorting_columns"`
}

type RowGroup struct{ pq.RowGroup }

func (g RowGroup) ToStat() BasicRowGroupStat {
	return BasicRowGroupStat{
		NumRows:        g.RowGroup.NumRows(),
		ColumnChunks:   int64(len(g.RowGroup.ColumnChunks())),
		SortingColumns: int64(len(g.RowGroup.SortingColumns())),
	}
}

type RowGroupsSource func() []pq.RowGroup

type ParquetFile struct{ *pq.File }

func (f ParquetFile) RowGroups() []pq.RowGroup { return f.File.RowGroups() }

func (f ParquetFile) AsRowGroupsSource() RowGroupsSource { return f.RowGroups }

func (s RowGroupsSource) ToStatIter() iter.Seq[BasicRowGroupStat] {
	return func(yield func(BasicRowGroupStat) bool) {
		var grps []pq.RowGroup = s()
		for _, grp := range grps {
			rgrp := RowGroup{RowGroup: grp}
			var stat BasicRowGroupStat = rgrp.ToStat()
			if !yield(stat) {
				return
			}
		}
	}
}

func BasicStatsToJsonWriter(
	stats iter.Seq[BasicRowGroupStat],
	enc *json.Encoder,
) error {
	for stat := range stats {
		e := enc.Encode(stat)
		if nil != e {
			return e
		}
	}

	return nil
}

type JsonWriter struct{ io.Writer }

func (w JsonWriter) WriteBasicStats(stats iter.Seq[BasicRowGroupStat]) error {
	var bw *bufio.Writer = bufio.NewWriter(w.Writer)
	var enc *json.Encoder = json.NewEncoder(bw)
	e := BasicStatsToJsonWriter(stats, enc)
	return errors.Join(e, bw.Flush())
}

func BasicStatsToJsonsToStdout(
	stats iter.Seq[BasicRowGroupStat],
) error {
	jw := JsonWriter{Writer: os.Stdout}
	return jw.WriteBasicStats(stats)
}

func (s RowGroupsSource) WriteAll(jw JsonWriter) error {
	var stats iter.Seq[BasicRowGroupStat] = s.ToStatIter()
	return jw.WriteBasicStats(stats)
}

func (s RowGroupsSource) StatsToJsonsToStdout() error {
	var stats iter.Seq[BasicRowGroupStat] = s.ToStatIter()
	return BasicStatsToJsonsToStdout(stats)
}

type ParquetFilename struct{ Filename string }

func (f ParquetFilename) ToRowGroupsSource() (RowGroupsSource, error) {
	file, err := os.Open(f.Filename)
	if nil != err {
		return nil, err
	}
	defer file.Close() //nolint:errcheck

	stat, err := file.Stat()
	if nil != err {
		return nil, err
	}

	var sz int64 = stat.Size()
	pfile, err := pq.OpenFile(file, sz)
	if nil != err {
		return nil, err
	}

	return ParquetFile{File: pfile}.AsRowGroupsSource(), nil
}

func (f ParquetFilename) StatsToJsonsToStdout() error {
	rgs, e := f.ToRowGroupsSource()
	if nil != e {
		return e
	}

	return rgs.StatsToJsonsToStdout()
}
