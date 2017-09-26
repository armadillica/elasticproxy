#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 new-version" >&2
    exit 1
fi

sed "s/elasticProxyVersion = \"[^\"]*\"/elasticProxyVersion = \"$1\"/" -i elasticproxy.go
sed "s/ELASTICPROXY_VERSION=\"[^\"]*\"/ELASTICPROXY_VERSION=\"$1\"/" -i docker/_version.sh

git diff
echo
echo "Don't forget to commit and tag:"
echo git commit -m \'Bumped version to $1\' elasticproxy.go docker/_version.sh
echo git tag -a v$1 -m \'Tagged version $1\'
