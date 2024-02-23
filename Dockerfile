# Use the official Golang image to create a build artifact.
FROM golang:1.22 AS build-env

@ Install the necessary dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install the latest version of Delve
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Copy the source code of your application inside the container
COPY . /app
WORKDIR /app

# Build the application
RUN go build -gcflags="all=-N -l" -o /server

# Final stage
FROM debian:bookworm

# Expose the port on which the debugger will listen
EXPOSE 39999 39999

# Copy the Delve binary and the server binary from the previous stage
WORKDIR /
COPY --from=build-env /go/bin/dlv /
COPY --from=build-env /server /

# Run the dlv server by default when the container starts
#ENTRYPOINT ["dlv", "debug", "--listen=:39999", "--headless=true", "--api-version=2", "--continue", "--accept-multiclient"]
CMD ["/dlv", "--listen=:39999", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/server"]