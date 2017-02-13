FROM alpine

MAINTAINER Denis Zheleztsov <difrex.punk@gmail.com>

# Usage:
# Build docker image:
# docker build -t lessmore_builder -f Dockerfile.builder .
# Build binary artifact:
# docker run -ti -v $(pwd)/out:/out/ lessmore_builder

# Install depends
RUN apk update && apk add git go

ENV GOPATH /usr

# Get sources
RUN cd / && git clone https://gitea.difrex.ru/Umbrella/lessmore.git

# Get go depends
RUN cd /lessmore && go get -t -v ./... || true
RUN cd /lessmore && go get gitea.difrex.ru/Umbrella/fetcher
RUN cd /lessmore && go get gitea.difrex.ru/Umbrella/lessmore

ENTRYPOINT cd /lessmore && go build && mv lessmore /out/
