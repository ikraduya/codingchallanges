#!/bin/bash

set -x

go build

./ccwc -c test.txt

./ccwc -l test.txt

./ccwc -w test.txt

wc -m test.txt
./ccwc -m test.txt

./ccwc test.txt

cat test.txt | ./ccwc -l
