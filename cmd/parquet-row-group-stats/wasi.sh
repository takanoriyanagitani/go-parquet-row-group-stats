#!/bin/sh

tinygo \
	build \
	-o ./parquet-row-group-stats.wasm \
	-target=wasip1 \
	-opt=z \
	-no-debug \
	./main.go
