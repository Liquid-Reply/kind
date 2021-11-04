# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# kind node base image
#
# For systemd + docker configuration used below, see the following references:
# https://systemd.io/CONTAINER_INTERFACE/

# start from ubuntu, this image is reasonably small as a starting point
# for a kubernetes node image, it doesn't contain much we don't need
ARG BASE_IMAGE=ubuntu:21.10
FROM $BASE_IMAGE as build

# `docker buildx` automatically sets this arg value
ARG TARGETARCH

# copy in static files
# all scripts are 0755: http://www.filepermissions.com/file-permission/0755
COPY --chmod=0755 files/usr/local/bin/* /usr/local/bin/
COPY --chmod=0755 files/etc/systemd/system/* /etc/systemd/system/


RUN echo "Installing Packages ..." \
    && DEBIAN_FRONTEND=noninteractive clean-install \
      systemd curl ca-certificates iproute2 \
    && find /lib/systemd/system/sysinit.target.wants/ -name "systemd-tmpfiles-setup.service" -delete \
    && rm -f /lib/systemd/system/multi-user.target.wants/* \
    && rm -f /etc/systemd/system/*.wants/* \
    && rm -f /lib/systemd/system/local-fs.target.wants/* \
    && rm -f /lib/systemd/system/sockets.target.wants/*udev* \
    && rm -f /lib/systemd/system/sockets.target.wants/*initctl* \
    && rm -f /lib/systemd/system/basic.target.wants/* \
    && echo "ReadKMsg=no" >> /etc/systemd/journald.conf \
    && ln -s "$(which systemd)" /sbin/init

RUN echo "Adjusting systemd-tmpfiles timer" \
    && sed -i /usr/lib/systemd/system/systemd-tmpfiles-clean.timer -e 's#OnBootSec=.*#OnBootSec=1min#'

# TODO(0xe282b0): just hacking krustlet and kind version into the image
# Better: create `kind build krustlet-image` that builds the krustlet from sources
RUN curl -sSL --retry 5 --output /tmp/krustlet.${TARGETARCH}.tar.gz "https://krustlet.blob.core.windows.net/releases/krustlet-canary-linux-amd64.tar.gz" \
    && tar -C /usr/local/bin -xzvf /tmp/krustlet.${TARGETARCH}.tar.gz krustlet-wasi \
    && rm -f /tmp/krustlet.${TARGETARCH}.tar.gz
RUN mkdir /kind && echo "v1.23.0-alpha.3" > /kind/version

# squash
FROM scratch
COPY --from=build / /
# tell systemd that it is in docker (it will check for the container env)
# https://systemd.io/CONTAINER_INTERFACE/
ENV container docker
# systemd exits on SIGRTMIN+3, not SIGTERM (which re-executes it)
# https://bugzilla.redhat.com/show_bug.cgi?id=1201657
STOPSIGNAL SIGRTMIN+3

ENTRYPOINT [ "/sbin/init" ]