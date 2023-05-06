FROM golang:1.19.4

WORKDIR /app

COPY go.mod go.sum ./

# Install dependencies
RUN go get -u github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/websocket
RUN go get gopkg.in/yaml.v2

# Copy the application code
COPY . .

# Build the application
RUN go build -o client .

# Set the command to run when the container starts
CMD ["./client", "--ip", "server", "--client_port", "9999"]