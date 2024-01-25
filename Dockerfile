FROM golang:1.21 as builder
WORKDIR /go/src/app
COPY . .
RUN go mod download
# We could even run tests here! Since there's only 1 :)
RUN CGO_ENABLED=0 go build -C cmd/objectstorage/ -o /go/bin/app

# Docker is used as a base image so you can easily start playing around in the container using the Docker command line client.
FROM gcr.io/distroless/static-debian12
COPY --from=builder /go/bin/app /
COPY config.yaml config.yaml
CMD [ "/app" ]
