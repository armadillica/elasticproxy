#!/bin/bash -e

source _version.sh
GID=$(id -g)

# Use Docker to get Go in a way that allows overwriting the
# standard library with statically linked versions.
docker run -i --rm \
    -v $(pwd):/docker \
    -v "${GOPATH}:/go-local" \
    --env GOPATH=/go-local \
     golang:1.9 /bin/bash -ex << EOT
CGO_ENABLED=0 go get -a -ldflags '-s' github.com/armadillica/elasticproxy
install -g $GID -o $UID --strip \${GOPATH}/bin/elasticproxy /docker
EOT

# Use the statically linked executable to build our final Docker image.
docker build -t armadillica/elasticproxy:${ELASTICPROXY_VERSION} .
docker tag armadillica/elasticproxy:${ELASTICPROXY_VERSION} armadillica/elasticproxy:latest
echo "Successfully tagged armadillica/elasticproxy:latest"
