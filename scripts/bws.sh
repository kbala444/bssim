#!/bin/bash
# used to run a workload with a bunch of differnet lat and bw combos
# argument should be workload file path

. common.sh

cd ..

bandwidth=( 5 10 15 20 25 30 35 50)
for i in ${bandwidth[@]}
do
    ./bssim -wl $1 -bw $i -lat 0
done

# change workload in config for graphing
# comma separators since $1 might have slashes in it
sed -i "s,workload=.*,workload=$1,g" ./data/config.ini
python ./data/grapher.py
