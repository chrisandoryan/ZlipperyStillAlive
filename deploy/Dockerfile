# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN chmod -R -w /app

RUN chmod +w /app/storage

# Expose port 8080 to the outside world
EXPOSE 9002

# Command to run the executable
CMD ["./main"]