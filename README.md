# messagerApp-Go

## Implementation Details
### Setup for Your Database
1. Create `config.yaml` in root in the format below:
```yaml
db:
  host: localhost
  port: YOUR_PORT_NUMBER
  user: USER_NAME
  password: YOUR_PASS_WORD
  name: DATABASE_NAME
```
2. Excute `script/createDatabase_chatApp.sql` and `createSamples_chatApp.sql` to set up your database (MySQL)

3. Install all the dependencies:
```bash
go get "github.com/go-sql-driver/mysql"
go get "github.com/gorilla/websocket"
go get "gopkg.in/yaml.v2"
```

### Launch Server
Go to `./server` directory, run the command below:
```bash
go run main.go server.go user.go
```
### Launch Client
```bash
go run client.go handlerDatabase.go
```

## Usage
