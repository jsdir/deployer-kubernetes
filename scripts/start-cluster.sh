#!/bin/sh
set -e

# Start a kubernetes cluster.
# llamashoes/dind-kubernetes
docker run -d -p 127.0.0.1:8888:8888 --privileged bzsarwilgcx774cgqbutbvk

# Block until the service starts.
echo "Blocking until server starts..."
wget --tries=30 --retry-connrefused --waitretry=1 http://localhost:8888 &> /dev/null
echo "Server at localhost:8888 started."
