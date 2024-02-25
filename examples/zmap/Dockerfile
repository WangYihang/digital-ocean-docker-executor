# syntax=docker/dockerfile:1

# Build zmap
FROM ubuntu:22.04

USER root

# Install dependencies
RUN apt-get update \
    && apt-get install -y git build-essential cmake libgmp3-dev gengetopt libpcap-dev flex byacc libjson-c-dev pkg-config libunistring-dev libjudy-dev \
    && apt-get autoremove -yqq --purge \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Build and install zmap
RUN cd /opt \
    && git clone https://github.com/zmap/zmap \
    && cd zmap \
    && cmake . \
    && make -j4 \
    && make install

# Define default command
ENTRYPOINT [ "zmap" ]