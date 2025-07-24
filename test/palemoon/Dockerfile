FROM golang:alpine AS build

RUN go install github.com/mattn/goreman@latest

FROM debian:latest AS run
RUN apt-get update -y && \
  apt-get install -y tigervnc-standalone-server dwm xterm && \
  rm -rf /var/lib/apt/lists
ENV VNC_DESKTOP_NAME="Xtigervnc"
ENV VNC_GEOMETRY=1280x800
COPY app/ /app/
COPY --from=build /go/bin/goreman /usr/bin/goreman

CMD ["/app/main.sh"]