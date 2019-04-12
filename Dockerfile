
# Multi stage build.
# First Create a build image and build the entire application here
FROM golang:latest
WORKDIR /go/src/gitlab.com/gilden/fortis/

# Copy the project
COPY . .

# Run the build script
RUN bash build/build.sh -linux

# Create an application image that will be pushed to re register
FROM google/debian:wheezy
MAINTAINER Remy Span

# Add the api binary
COPY --from=0 /go/src/gitlab.com/gilden/fortis/bin .

ENV PORT 6767
EXPOSE 6767
CMD ["./fortis"]