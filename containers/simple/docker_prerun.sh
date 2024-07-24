#!/bin/sh


function image_exists_p () {
    # check for an image matching a specific tag
    local tag=$1
    docker inspect $tag > /dev/zero
}

function docker_prerun () {
    local tag=$1
    local context=$2
    if [ -z ${tag+x} ]; then   # TAG is not set
        echo "Must pass tag as first arg"
        exit -1
    fi

    if ! [[ -d $context ]]; then
        echo "Must pass context directory as second arg."
        exit -1
    fi

    image_exists_p $tag
    if [ "$?" -eq 0 ]; then   # nothing more to do
        echo "image_exists_p found an existing image with $tag"
        return 0
    else
        cd $context
        tar cz --file mydate.tgz ./usr ./lib64
        docker import mydate.tgz
        IMAGE_ID=$(docker images | awk 'NR > 1 && NR < 3 {print $3}')
        docker tag $IMAGE_ID $tag
    fi
    return 0
}

if [ "2" -ne "$#" ]; then
    printf "usage: %s TAG CONTEXT\n" "$0"
    exit -1
fi

docker_prerun "$@"
