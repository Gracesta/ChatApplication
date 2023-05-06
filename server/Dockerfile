# server
FROM golang:1.19.4

# Set the working directory
WORKDIR /app/server


COPY go.mod go.sum ./



# Copy the application code
COPY . .

ENV PORT=8888

EXPOSE 8888

# # Build the application
RUN go build -o server .

# Set the command to run when the container starts
CMD ["./server", "--ip", "0.0.0.0"]