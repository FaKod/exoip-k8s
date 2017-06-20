#FROM debian:jessie
FROM alpine:3.3
MAINTAINER Christopher Schmidt <fakod666@gmail.com>

# so apt-get doesn't complain
ENV DEBIAN_FRONTEND=noninteractive

#RUN \
#  apt-get update && \
#  apt-get install -y ca-certificates && \
#  rm -rf /var/lib/apt/lists/*

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ADD exoip-k8s exoip-k8s
ADD run.sh /run.sh
ENTRYPOINT ["/run.sh"]
