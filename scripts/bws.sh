#!/bin/bash
. common.sh
cd ..
for i in "${bws[@]}"
do
    echo $i
    replaceOpt $1 "bandwidth" $i
    ./bssim ./$1
done

