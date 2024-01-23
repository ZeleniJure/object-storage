FROM golang:1.15 as builder
WORKDIR /mnt/app
COPY . .
RUN go build

# Docker is used as a base image so you can easily start playing around in the container using the Docker command line client.
FROM docker
RUN apk add bash curl
COPY --from=builder /mnt/app/object-storage /usr/local/bin/object-storage
