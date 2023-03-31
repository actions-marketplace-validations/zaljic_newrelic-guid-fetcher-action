# Use the official Golang image to create a build artifact.
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Create appuser.
RUN adduser -D -g '' appuser

# Set the current working directory inside the container.
WORKDIR /app

# Copy go mod and sum files.
COPY . /app

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed.
RUN go get -d -v

# Build the Go app. CGO_ENABLED=0 is required to build a static executable. 
# -ldflags="-w -s" is used to reduce the size of the executable.
# -v is used to show the build progress.
# -o is used to specify the output file name.
# . is used to specify the current directory as the source.
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

# A distroless container image with some basics like SSL certificates.
# https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static

# Copy the Pre-built binary file from the previous stage
# and set it as the entrypoint of the container.
COPY --from=builder /app/app /app
ENTRYPOINT ["/app"]
