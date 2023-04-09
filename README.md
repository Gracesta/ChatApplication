# ChatApplication-Go
This project is adopted following languages:<br>
`Golang` (backend), `HTML+JavaScript+CSS` (frontend), `MySQL` (database) <br>
And Following technologies:<br>
`Ajax` (Frontend to backend), `Websocket` (Backend to Frontend) <br>

## Update
[2023/04/09] First time released to public

## Implementation Details
### Setup for Your Database
1. Create `config.yaml` in root in the format below:
```yaml
db:
  host: localhost
  port: YOUR_PORT_NUMBER
  user: USER_NAME
  password: YOUR_PASSWORD
  name: DATABASE_NAME
```
2. Navigate to `./scripts/` and excute `createDatabase_chatApp.sql` and `createSamples_chatApp.sql` to set up your database (MySQL)

3. Install all the dependencies:
```bash
go get "github.com/go-sql-driver/mysql"
go get "github.com/gorilla/websocket"
go get "gopkg.in/yaml.v2"
```

### Launch Server
Navigate to `./server` directory, run the command below:
```bash
go run main.go server.go user.go
```
### Launch Client
```bash
go run client.go handlerDatabase.go
```

Finally navigate to the showed page to login (e.g.):
```bash
Visit Page on: http://localhost:XXXX/
```

## Demo

### Login Page
![Login](./imgs/login.jpg "Login")

### Register Page
![Register](./imgs/register.jpg "Register")

### Homepage (Only Public chat mode to select now)
![Homepage](./imgs/homePage.jpg "Homepage")

### Group Chat Page
![Chat Page](./imgs/chatSingle.jpg "Chat Page")

![Group Chat Page](./imgs/chat_demo.jpg "Group Chat Page")

