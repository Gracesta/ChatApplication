# server
FROM golang:1.19.4

WORKDIR /app/server

COPY go.mod go.sum ./
COPY . .

ENV PORT=8888
EXPOSE 8888

# # Build the application
RUN go build -o server .

# Set the command to run when the container starts
CMD ["./server", "--ip", "0.0.0.0"]