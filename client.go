package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

type Client struct {
	Ip   string
	Port int
	Name string
	flag int

	conn    net.Conn
	db      *sql.DB
	user_id int
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		Ip:   ip,
		Port: port,
		flag: 999,
	}

	// Link to server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn

	return client
}

func (client *Client) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to WebSocket
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Loop to read data from client.conn and send to frontend
	for {
		// Read data from client.conn
		buf := make([]byte, 1024)
		n, err := client.conn.Read(buf)
		if err != nil {
			break
		}
		msg := string(buf[:n])
		fmt.Println("Receive Message in websocket from user conn:", msg)
		// Handle message here
		// msgStrings := strings.Split(msg, "|")[1:]
		// msg[]
		// msgUserName := msgStrings[1]
		// msgContent := msgStrings[2]
		// fmt.Println("Receive Message in websocket from user conn:", msgAddr, msgContent)

		// Send data to frontend over WebSocket
		if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			break
		}
	}
	// Close WebSocket connection
	ws.Close()
}

func checkLogin(db *sql.DB, username, password string) (int, error) {
	var user_id int
	err := db.QueryRow("SELECT user_id FROM users WHERE username=? AND password=?", username, password).Scan(&user_id)
	if err != nil {
		return user_id, err
	}
	if user_id != 0 {
		return user_id, nil
	}
	return user_id, nil
}

func (client *Client) loginVerificationHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	// fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// fmt.Println("json data", data)
	// Get the value of the input_message field from the request body

	username, ok := data["username"]
	password, ok := data["password"]
	if !ok {
		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
		return
	}
	var user_id int

	// checkLogin function will return 'user_id' and 'error'
	user_id, err = checkLogin(client.db, username, password)
	if err != nil {
		fmt.Println("Failed to login with database: %v", err)
	}

	// Verification on login process
	var verf string
	if user_id != 0 {
		client.user_id = user_id
		fmt.Println("Login succeeded, user_id is:", user_id)
		verf = "TRUE"
		// chat application homepage
		go http.HandleFunc("/chat", client.chatHandler)

		// WebSocket endpoint for the chat messages
		go http.HandleFunc("/ws", client.handleWebSocket)

		// send message
		go http.HandleFunc("/send-message", client.handleSendMessage)
	} else {
		verf = "FALSE"
	}

	// TODO: pass data from database to other handlers

	// Send a response back to the client
	response := map[string]string{"status": "ok", "verification": verf}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (client *Client) Run() {
	// Serve static files on website
	http.Handle("/", http.FileServer(http.Dir("static")))

	// launch database for client
	type Config struct {
		Database struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Name     string `yaml:"name"`
		} `yaml:"db"`
	}

	configFile := "config.yaml"

	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Failed to open settings file: %v", err)
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode settings: %v", err)
	}

	// Connect to MySQL database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	client.db = db

	log.Println("Database launched", db)

	// Here for login process
	http.HandleFunc("/login-verif", client.loginVerificationHandler)
	// // chat application homepage
	// go http.HandleFunc("/chat", client.chatHandler)

	// // WebSocket endpoint for the chat messages
	// go http.HandleFunc("/ws", client.handleWebSocket)

	// // send message
	// go http.HandleFunc("/send-message", client.handleSendMessage)
	// // WebSocket endpoint for the chat messages
	// go http.HandleFunc("/receive-message", client.handleReceivedMessage)

	// Get a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	portAddr := fmt.Sprintf(":%d", port)

	// Start the server on port 8080

	log.Println("Starting server on port: ", port)
	log.Printf("Visit Page on: http://localhost:%d/", port)
	// err := http.ListenAndServe(":8082", nil)
	err = http.ListenAndServe(portAddr, nil)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

	for {

	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Server IP")
	flag.IntVar(&serverPort, "port", 8888, "Server Port")
}

func main() {
	// command line parse
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> Link to Server Falied <<<<<")
		return
	}

	fmt.Println(">>>>> Link to Server Succedded <<<<<")

	client.Run()

}

type Chatlog struct {
	Username  string
	Content   string
	Timestamp string
}

func loadChatLogsFromDatabase(client *Client) struct {
	Tile     string
	Chatlogs []Chatlog
} {
	db := client.db
	rows, err := db.Query("SELECT username, message, timestamp FROM users u JOIN chat_logs cl ON u.user_id = cl.user_id")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	data := struct {
		Tile     string
		Chatlogs []Chatlog
	}{}

	for rows.Next() {
		var username, chatlog, timestamp string
		err := rows.Scan(&username, &chatlog, &timestamp)

		// Check for errors when scanning the query results
		if err != nil {
			log.Fatal(err)
		}

		data.Chatlogs = append(data.Chatlogs, Chatlog{Username: username, Content: chatlog, Timestamp: timestamp})
	}
	return data
}

func (client *Client) chatHandler(w http.ResponseWriter, r *http.Request) {
	// Compile the chat template
	fmt.Println("chatHandler")
	tmpl := template.Must(template.ParseFiles("./static/chat.html"))

	// Render the template with any necessary data
	// TODO: load the chat logs in database to chatwindow
	// data := struct {
	// 	Title    string
	// 	ChatLogs []string
	// }{
	// 	Title:    "Group Chat Application",
	// 	ChatLogs: []string{"m_1", "m_2"},
	// }
	data := loadChatLogsFromDatabase(client)

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func timesStampMySQLFormat(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		panic(err.Error())
	}

	// Format the time.Time object into MySQL-compatible format
	mysqlTimestamp := t.Format("2006-01-02 15:04:05")
	return mysqlTimestamp
}

func (client *Client) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as JSON
	fmt.Println("client.handle send message:")

	var data map[string]string
	// fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// fmt.Println("json data", data)
	// Get the value of the input_message field from the request body
	chatMsg, ok := data["input_message"]
	timestamp, ok := data["timestamp"]
	fmt.Println(timestamp)
	if !ok {
		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
		return
	}

	// Do something with the input message, such as send it to other users in the chat
	fmt.Println("Broadcast message to other users:", chatMsg)
	if len(chatMsg) != 0 {
		sendMsg := []byte(chatMsg + "\n")
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("client conn.write error: ", err)
		}

	}
	// INSERT this message to database
	stmt, err := client.db.Prepare("INSERT INTO chat_logs(user_id, message, timestamp) VALUES(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec(client.user_id, chatMsg, timesStampMySQLFormat(timestamp))
	if err != nil {
		log.Fatal(err)
	}

	// Print the number of rows affected by the insert statement
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d row(s) inserted.\n", rowsAffected)

	// Send a response back to the client
	message_backend := "From backend"
	response := map[string]string{"status": "ok", "message_backend": message_backend}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (client *Client) SelectUser() {

	// send who to know the user online
	_, err := client.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("client conn.write error: ", err)
		return
	}
}

func (client *Client) PrivateChat() {

	// show online user first
	client.SelectUser()

	var targetUserName string
	var chatMsg string

	fmt.Println(">>>>> Input the name of user who you wanna chat with")
	fmt.Scanln(&targetUserName)

	for targetUserName != "exit" {
		var errBufio error

		fmt.Println(">>>>> Input the message you wanna send (input \"exit\" to exit)")
		in := bufio.NewReader(os.Stdin)
		chatMsg, errBufio = in.ReadString('\n')
		if errBufio != nil {
			fmt.Println("reading string error", errBufio)
		}
		// fmt.Scanln(&chatMsg)

		// send message to usr conn to broadcast
		for chatMsg != "exit\r\n" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + targetUserName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("client conn.write error: ", err)
					break
				}

				chatMsg = ""
				fmt.Println(">>>>> Input the message you wanna send (input \"exit\" to exit)")
				chatMsg, errBufio = in.ReadString('\n')
				if errBufio != nil {
					fmt.Println("reading string error", errBufio)
				}
				// fmt.Scanln(&chatMsg)
			}
		}

		client.SelectUser()
		fmt.Println(">>>>> Input the name of user who you want chat with (input \"exit\" to exit)")
		fmt.Scanln(&targetUserName)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("Input the new username you want")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn.write error: ", err)
		return false
	}

	return true

}

// func (client *Client) menu() bool {
// 	var flag int

// 	fmt.Println("1. Public Chat")
// 	fmt.Println("2. Private Chat")
// 	fmt.Println("3. Update Username")
// 	fmt.Println("0. Exit")

// 	fmt.Scanln(&flag)

// 	if flag >= 0 && flag <= 3 {
// 		client.flag = flag
// 		return true
// 	} else {
// 		fmt.Println("Input number within range")
// 		return false
// 	}

// }
