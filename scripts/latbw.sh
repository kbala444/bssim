#!/bin/bash
# used to run a workload with a bunch of differnet lat and bw combos
# argument should be workload file path

. common.sh

cd ..

for i in ${latency[@]}
do
    for j in ${bandwidth[@]}
    do
	./bssim -wl $1 -bw $j -lat $i
    done
done

# change workload in config for graphing
# comma separators since $1 might have slashes in it
sed -i "s,workload=.*,workload=$1,g" ./data/config.ini
python ./data/grapher.py
