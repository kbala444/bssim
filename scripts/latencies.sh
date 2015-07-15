#!/bin/bash
# used to run a workload with a bunch of differnet lat and bw combos
# argument should be workload file path

. common.sh

cd ..

latency=( 2 19 27 23 38 58 )

for i in ${latency[@]}
do
    ./bssim -wl $1 -lat $i -bw 1000
done

# change workload in config for graphing
# comma separators since $1 might have slashes in it
sed -i "s,workload=.*,workload=$1,g" ./data/config.ini
python ./data/grapher.py
