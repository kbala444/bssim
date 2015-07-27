#!/bin/bash
cd ..
outfile='metrics/data'
workload='samples/viral'
#outfile=$1

printf "\n$workload\n" >> $outfile
for i in `seq 1 20`;
do 
    stats="$(./bssim -wl $workload | tail -10)" 
    meanbt=$(echo "$stats" | awk -F':' '/Mean\ block/ {print $2}' | tr -d ' ')
    breceived=$(echo "$stats" | awk -F':' '/Total\ blocks/ {print $2}' | tr -d ' .')
    dupb=$(echo "$stats" | awk -F':' '/Duplicate\ blocks/ {print $2}' | tr -d ' .')
    # trim last char (a period) from meanbt
    echo "$i, ${meanbt%?}, $breceived, $dupb" >> $outfile
done
