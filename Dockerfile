# Dockerfile References: https://docs.docker.com/engine/reference/builder/
# Start from the latest golang base image
FROM golang:1.24.0 as builder

ARG packagePath
ARG buildNum
ARG circleSha
ARG TARGETOS
ARG TARGETARCH

# Set the Current Working Directory inside the container
WORKDIR /app

# Set git credentials for private repo access
# RUN git config --global credential.helper store && echo "${DOCKER_GIT_CREDENTIALS}" > ~/.git-credentials

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Emit debug info
RUN echo "Building for OS: $TARGETOS, Arch: $TARGETARCH"
RUN echo "Package Path: $packagePath"
RUN echo "Build Number: $buildNum"
RUN echo "Commit SHA: $circleSha"

# Build the Go app for the correct architecture
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -ldflags "-X ${packagePath}/version.BuildNumber=${buildNum} -X ${packagePath}/version.CommitID=${circleSha} -X '${packagePath}/version.Prerelease=-'" -installsuffix cgo -o main ./

######## Start a new stage from scratch #######
FROM --platform=$TARGETPLATFORM ubuntu:24.04

# Specialized tools for package-repo
RUN apt update
RUN apt upgrade -y
RUN apt install -y git gnupg dpkg-dev apt-utils nano

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main", "start"]