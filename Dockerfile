FROM golang:1.21 as build
WORKDIR /build
COPY go.mod main.go ./
COPY upload.html done.html ./
ENV CGO_ENABLED=0
RUN go build

FROM debian:12-slim
LABEL org.opencontainers.image.source https://github.com/sikalabs/filedrop
RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends ca-certificates && \
  rm -rf /var/lib/apt/lists/*
COPY --from=build /build/filedrop /usr/local/bin/filedrop
CMD ["/usr/local/bin/filedrop"]
EXPOSE 8000
