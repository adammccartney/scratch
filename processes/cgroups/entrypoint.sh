#!/bin/bash

timeout ${1}s coro-sieve > /dev/null &
while :
do 
    unset nproc
    nproc=$(pstree -p $! | wc -l)
    sleep 1
    if [ $nproc == "0" ]; then
        exit 0
    else
        echo -e $(date) -- INFO -- coro-sieve forked $nproc processes
    fi
done
