#!/bin/bash

set -x

make

./ccgrep "" test.txt | diff test.txt -
echo ""

./ccgrep J rockbands.txt
echo ""

./ccgrep -r Nirvana *
echo ""

./ccgrep -r Nirvana * | ./ccgrep -v Madonna
echo ""

./ccgrep "[[:digit:]]" test-subdir/BFS1985.txt
echo ""

./ccgrep "\w" symbols.txt
echo ""

./ccgrep ^A rockbands.txt
echo ""

./ccgrep na$ rockbands.txt
echo ""

./ccgrep A rockbands.txt | wc -l
echo ""

./ccgrep -i A rockbands.txt | wc -l
echo ""
