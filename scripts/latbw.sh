#!/bin/bash
latencies=( 0 1 2 3 5 8 10 15 25 40 )
bws=( 50 100 150 200 500 1000 5000 10000 )

cd ..
for i in "${latencies[@]}"
do
    for j in ${bws[@]}
    do
	sed -i "s/latency:[ ]*[0-9]\+/latency: $i/g" samples/star
	sed -i "s/bandwidth:[ ]*[0-9]\+/bandwidth: $j/g" samples/star
	./bssim ./samples/star
    done
done

