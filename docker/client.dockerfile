ARG PLATFORM=amd64
FROM ${PLATFORM}/golang:1.23.4-bookworm AS builder

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -qq -y install \
  curl \
  gstreamer1.0* \
  wget \
  build-essential \
  pkg-config \
  libgstreamer1.0-dev \
  libgstreamer-plugins-base1.0-dev \
  libgstreamer-plugins-bad1.0-dev \
  unzip

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -qq -y install \
  git


RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY .. /opt/app/
#COPY Taskfile.yml /opt/app/Taskfile.yml

WORKDIR /opt/app
RUN go build .

FROM ${PLATFORM}/debian:bookworm

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -qq -y install \
    gstreamer1.0*

COPY --from=builder /opt/app/pion-webrtc /opt/app/pion-webrtc
COPY --from=builder /opt/app/static /opt/app/static
WORKDIR /opt/app
ENV WEBRTC_SERVER_LOCATION=ws://localhost:8080/ws
CMD ["/opt/app/pion-webrtc", "client"]