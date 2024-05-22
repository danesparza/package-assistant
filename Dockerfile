# Dockerfile References: https://docs.docker.com/engine/reference/builder/
# Start from the latest golang base image
FROM golang:1.22.2 as builder

ARG packagePath
ARG buildNum
ARG circleSha

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

# Emit what we have
RUN echo $packagePath
RUN echo $buildNum
RUN echo $circleSha

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-X ${packagePath}/version.BuildNumber=${buildNum} -X ${packagePath}/version.CommitID=${circleSha} -X '${packagePath}/version.Prerelease=-'" -installsuffix cgo -o main ./

######## Start a new stage from scratch #######
FROM ubuntu:23.10

# Specialized tools for package-repo
RUN apt update
RUN apt upgrade
RUN apt install gnupg dpkg-dev apt-utils nano

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main", "start"]