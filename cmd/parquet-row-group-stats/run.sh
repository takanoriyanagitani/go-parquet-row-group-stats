#!/bin/sh

genschema(){
	echo '
		message spark_schema {
			REQUIRED BINARY timestamp(UTF8);
			REQUIRED BINARY severity(UTF8);
			REQUIRED BINARY body(UTF8);
		}
	'
}

gencsv0(){
	echo timestamp,severity,body
	echo 2025-10-03T08:24:17.012345Z,INFO,apt update done
	echo 2025-10-02T08:24:17.012345Z,WARN,apt update failure
}

gencsv1(){
	echo timestamp,severity,body
	echo 2024-10-03T08:24:17.012345Z,INFO,apt update done
	echo 2023-10-02T08:24:17.012345Z,WARN,apt update failure
	echo 2022-10-02T08:24:17.012345Z,WARN,apt update failure
}

geninput(){
	echo creating input files...

	mkdir -p ./sample.d

	genschema > ./sample.d/tmp.schema

	gencsv0 > ./sample.d/tmp0.csv
	gencsv1 > ./sample.d/tmp1.csv

	which parquet-fromcsv | fgrep -q parquet-fromcsv || exec sh -c '
		echo parquet-fromcsv missing.
		echo you can install it using cargo install parquet --bin ...
		exit 1
	'

	parquet-fromcsv \
		--has-header \
		--schema ./sample.d/tmp.schema \
		--input-file ./sample.d/tmp0.csv \
		--output-file ./sample.d/input0.parquet

	parquet-fromcsv \
		--has-header \
		--schema ./sample.d/tmp.schema \
		--input-file ./sample.d/tmp1.csv \
		--output-file ./sample.d/input1.parquet
}

test -f ./sample.d/input0.parquet || geninput
test -f ./sample.d/input1.parquet || geninput

run_native(){
	ENV_PARQUET_FILENAME=./sample.d/input0.parquet ./parquet-row-group-stats 
	ENV_PARQUET_FILENAME=./sample.d/input1.parquet ./parquet-row-group-stats 
}

run_wazero(){

	wazero \
		run \
		-mount "${PWD}/sample.d:/guest.d:ro" \
		-env ENV_PARQUET_FILENAME=/guest.d/input0.parquet \
		./parquet-row-group-stats.wasm

	wazero \
		run \
		-mount "${PWD}/sample.d:/guest.d:ro" \
		-env ENV_PARQUET_FILENAME=/guest.d/input1.parquet \
		./parquet-row-group-stats.wasm

}

run(){
	which wazero | fgrep -q wazero || run_native
	which wazero | fgrep -q wazero && run_wazero
}

run
