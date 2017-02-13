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

# Get go depends
RUN go get gitea.difrex.ru/Umbrella/fetcher
RUN go get gitea.difrex.ru/Umbrella/lessmore
RUN go install gitea.difrex.ru/Umbrella/lessmore

# Check build result
RUN echo -ne "Check build result\n==============="
RUN /usr/bin/lessmore --help || [[ $? -eq 2 ]]

ENTRYPOINT mv /usr/bin/lessmore /out/
