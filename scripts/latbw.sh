#!/bin/bash
# used to run a workload with a bunch of differnet lat and bw combos
# argument should be workload file path

. common.sh

cd ..

for i in "${latencies[@]}"
do
    for j in ${bws[@]}
    do
	replaceOpt $1 "latency" $i
	replaceOpt $1 "bandwidth" $j

	#cat $1
	./bssim ./$1
    done
done



