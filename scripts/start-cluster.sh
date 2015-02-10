#!/bin/sh
set -e

# Start a kubernetes cluster.
docker run -d --privileged --net="host" llamashoes/dind-kubernetes

# Block until the service starts.
wget --tries=30 --retry-connrefused --waitretry=1 http://localhost:8888 &> /dev/null
