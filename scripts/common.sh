#latency=( 0 10 20 30 40 50 60 70 )
#bandwidth=( 1 5 10 15 25 30 40 50 100 )

line=$(grep latencies= ../data/config.ini| cut -d "=" -f 2)
latencies=( $line )

line=$(grep bandwidths= ../data/config.ini| cut -d "=" -f 2)
bandwidths=($line)

replaceOpt(){
    echo $1
    echo $2
    echo $3
    if head -1 $1 | grep -q $2:; then
	sed -i "s/$2:[ ]*[0-9]\+/$2: $3/g" $1
    else
	sed -i "1 s/$/, $2: $3/" $1
    fi
}
