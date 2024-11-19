#!/bin/bash

ESC='\e[1A\e[K'

timeout ${1}s coro-sieve > /dev/null &
while :
do 
    unset num_children
    num_children=$(pstree -p $! | wc -l)
    sleep 1
    if [ $num_children == "0" ]; then
        exit 0
    else
        echo -e ${ESC} $(date) coro-sieve end with $num_children children
    fi
done
