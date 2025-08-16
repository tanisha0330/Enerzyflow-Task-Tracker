# Use an official Go image as the base
FROM golang:1.21-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY go.mod ./
COPY go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of your Go source code into the container
COPY . .

# Build the Go application into a single executable file
RUN go build -o /task-tracker-app

# This command will be run when the container starts
CMD [ "/task-tracker-app" ]