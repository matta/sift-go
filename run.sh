#!/bin/sh
set -x
echo ==== >./sift.log
SIFT_LOGFILE=sift.log go1.23rc1 run ./sift.go
