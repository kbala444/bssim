#!/bin/bash
bws=( 0 1 2 3 5 8 10 15 25 40 )
cd ..
for i in "${bws[@]}"
do
    echo $i
    sed -i "s/bandwidth:[ ]*[0-9]\+/bandwidth: $i/g" samples/star
    ./bssim samples/star
done

