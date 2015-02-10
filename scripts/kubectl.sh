#!/bin/sh
docker run -it --net='host' fandekasp/kube ctl --server='http://localhost:8888'
