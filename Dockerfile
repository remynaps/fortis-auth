
# Multi stage build.
# First Create a build image and build the entire application here
FROM golang:latest
WORKDIR /go/src/gitlab.com/gilden/fortis/

# Copy the project
COPY . .

# Fetch the dependencies
RUN go get -u github.com/golang/dep/...
RUN dep ensure

# Run the build script
RUN bash build/build.sh -linux

# Create an application image that will be pushed to re register
FROM google/debian:wheezy
MAINTAINER Remy Span

# Add the api binary
COPY --from=0 /go/src/gitlab.com/gilden/fortis/bin .
COPY ./config ./config

ENV PORT 6767
EXPOSE 6767
CMD ["./fortis_api"]