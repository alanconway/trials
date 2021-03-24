#!/bin/bash

GOBIN=${GOBIN:-$HOME/go/bin}
DIR=/tmp/symnotify
go install -v ../...
rm -rf $DIR
mkdir -p $DIR

pkill write-metric
$GOBIN/write-metric $DIR &
sleep 1
curl http://localhost:2112/metric
$GOBIN/stress
kill -int %%
