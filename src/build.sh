#!/bin/bash
REGISTRY=$1
VER=`git describe --tag`
TAG=$REGISTRY/proxy:$VER
cd .. && docker build -t $TAG \
    --build-arg version=$VER \
    --build-arg commit=`git rev-parse --short HEAD` \
    --build-arg date="`date --rfc-3339=seconds`" \
    .
docker push $TAG
