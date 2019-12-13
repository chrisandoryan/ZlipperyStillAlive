# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Put the flag
RUN mv flagahaha.txt /etc/spartan

RUN chmod -R -w /app

RUN chmod +w /app/storage

# Give some fun limit for the user :)
RUN chmod a-x /bin/cat
RUN chmod a-x /usr/bin/tail
RUN chmod a-x /usr/bin/head
RUN chmod a-x /bin/more
RUN passwd -l root
RUN useradd admin -m -d /home/admin -s /bin/rbash
USER admin
ENV HOME /home/admin

# Expose port 9002 to the outside world
EXPOSE 9002

# Command to run the executable
CMD ["./main"]
