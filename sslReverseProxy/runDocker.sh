#!/bin/bash
cd "$(dirname "$0")"

docker inspect sslRevProxy >/dev/null 2>&1 || docker build . -t sslrevproxy
# attaching pseudo terminal in interactive mode was only fix I could get working, --init also failed
docker run --rm -v ./:/var/run/sslRevProxy -p 8443:8443 sslrevproxy
