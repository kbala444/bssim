latency=( 0 6 9 11 13 ) #18 22 27 30 )
bandwidth=( 1000 5000 500 50 10000 )

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
