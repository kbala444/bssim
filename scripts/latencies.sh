#!/bin/bash
latencies=( 0 1 2 3 5 8 10 15 25 40 )
cd ..
for i in "${latencies[@]}"
do
    echo $i
    sed -i "s/latency:[ ]*[0-9]\+/latency: $i/g" samples/star
    ./bssim samples/star
done

