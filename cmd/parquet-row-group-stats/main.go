package main

import (
	"log"
	"os"

	rs "github.com/takanoriyanagitani/go-parquet-row-group-stats"
)

var filename string = os.Getenv("ENV_PARQUET_FILENAME")
var parquetFilename rs.ParquetFilename = rs.ParquetFilename{
	Filename: filename,
}

func main() {
	e := parquetFilename.StatsToJsonsToStdout()
	if nil != e {
		log.Fatal(e)
	}
}
