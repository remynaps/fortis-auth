
# Multi stage build.
# First Create a build image and build the entire application here
FROM golang:latest

RUN mkdir /fortis
WORKDIR /fortis

COPY ./go.mod ./go.sum ./

RUN go mod download

# Copy the project
COPY . .

# Run the build script
RUN bash build/build.sh -linux

# Create an application image that will be pushed to re register
FROM alpine:3.7
MAINTAINER Remy Span

# add openssl and trusted certificates
RUN apk add openssl ca-certificates

# Add the api binary
COPY --from=0 /fortis/bin .

# Add the required directories
COPY --from=0 /fortis/config ./config
COPY --from=0 /fortis/migrations ./migrations
COPY --from=0 /fortis/static ./static
COPY --from=0 /fortis/templates ./templates

ENV PORT 8081
EXPOSE 8081
CMD ["./fortis_api"]