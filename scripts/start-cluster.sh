#!/bin/sh
set -e

# Start a kubernetes cluster.
docker run --privileged --net="host" llamashoes/dind-kubernetes
