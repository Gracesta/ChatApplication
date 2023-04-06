# messagerApp-Go

## Implementation details
### Setup for your database
Create config.yaml in root in the format below:
db:
  host: localhost
  port: YOUR_PORT_NUMBER
  user: USER_NAME
  password: YOUR_PASS_WORD
  name: DATABASE_NAME
### Launch the server
go run main.go server.go user.go

### Launch client
go run client.go

## Usage
