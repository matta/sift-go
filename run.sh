#!/bin/sh
set -x
echo ==== > ./sift.log; SIFT_LOGFILE=sift.log go run ./sift.go
