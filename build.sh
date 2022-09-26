#!/bin/bash

BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')

go build -o sensor -ldflags "-X \"main.builddate=$BUILDDATE\"" main2.go
