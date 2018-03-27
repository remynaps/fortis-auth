# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:latest

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/remynaps/author

ADD ./content ~/content

# Add the application inside the container.
RUN go get github.com/remynaps/author
RUN go install github.com/remynaps/author
RUN go build -o author github.com/remynaps/author

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/author

# Document that the service listens on port 8080.
EXPOSE 8080
