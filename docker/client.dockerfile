ARG PLATFORM=amd64
FROM ${PLATFORM}/golang:1.23.4-alpine AS builder

RUN apk update && apk add --no-cache \
  curl \
  unzip

RUN apk update && apk add --no-cache \
  gstreamer-dev \
  gst-plugins-base-dev \
  gst-plugins-bad-dev

RUN apk add --no-cache git make build-base

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY .. /opt/app/
#COPY Taskfile.yml /opt/app/Taskfile.yml

WORKDIR /opt/app
RUN go build .

FROM ${PLATFORM}/alpine

RUN apk update && apk add --no-cache \
  gst-vaapi \
  gstreamer \
  gst-plugins-good \
  gstreamer-tools

COPY --from=builder /opt/app/pion-webrtc /opt/app/pion-webrtc
COPY --from=builder /opt/app/static /opt/app/static
WORKDIR /opt/app
ENV WEBRTC_SERVER_LOCATION=ws://localhost:8080/ws
CMD ["/opt/app/pion-webrtc", "client"]