#!/bin/bash

timeout ${1}s coro-sieve > /dev/null &
while :
do 
    unset num_children
    num_children=$(pstree -p $! | wc -l)
    sleep 1
    if [ $num_children == "0" ]; then
        exit 0
    else
        echo -e $(date) -- INFO -- coro-sieve forked $num_children children
    fi
done
