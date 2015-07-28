#!/bin/bash
# used to run single workload and show graphs for it
# argument should be filename

./bssim -wl $1

# change workload in config for graphing
# comma separators since $1 might have slashes in it
cd data
sed -i "s,workload=.*,workload=$1,g" ./config.ini
python grapher.py metrics config.ini
