#!/bin/bash
. common.sh
cd ..
for i in "${latencies[@]}"
do
    echo $i
    replaceOpt $1 "latency" $i
    ./bssim ./$1
done

