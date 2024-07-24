#!/bin/sh

function docker_run () {
    local tag=$1
    local cname=$2
    if [ -z ${tag+x} ]; then
        echo "Must pass a tag as first arg"
        exit -1
    fi
    if [ -z ${cname+x} ]; then
        echo "Must pass a tag as first arg"
        exit -1
    fi
    docker run --name ${cname} ${tag} dateloop
}

if [ "2" -ne "$#" ]; then
    printf "usage: %s TAG CONTAINER_NAME\n" "$0"
    exit -1
fi

docker_run "$@"
