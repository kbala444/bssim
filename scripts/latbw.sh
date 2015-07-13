#!/bin/bash
# used to run a workload with a bunch of differnet lat and bw combos
# argument should be workload file path

. common.sh

cd ..

for i in ${latency[@]}
do
    for j in ${bandwidth[@]}
    do
	replaceOpt $1 "latency" $i
	replaceOpt $1 "bandwidth" $j

	#cat $1
	./bssim ./$1
    done
done

# change workload in config for graphing
# comma separators since $1 might have slashes in it
sed -i "s,workload=.*,workload=$1,g" ./data/config.ini
sleep 1
python ./data/grapher.py
