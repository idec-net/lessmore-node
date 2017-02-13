FROM alpine

# Usage:
# Required artifact from lessmore_builder in $(pwd)/out/
# Build:
# docker build -t lessmore -f Dockerfile.app .
# Run:
# docker run -ti -p 15582:15582 lessmore -listen 0.0.0.0:15582 -es http://$ES:9200 -esindex idec -estype idec

MAINTAINER Denis Zheleztsov <difrex.punk@gmail.com>

ADD out/lessmore /usr/bin

ENTRYPOINT ["lessmore"]
