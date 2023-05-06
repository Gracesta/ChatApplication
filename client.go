package main

import (
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
	// server name
	// When running on Docker, the "ip" parameter will be set to server container name :"server"
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

		// Send data to frontend over WebSocket
		if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			break
		}
	}
	// Close WebSocket connection
	ws.Close()
}
func (client *Client) loginVerificationHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
		client.Name = username
		client.UpdateName() // retrieve username to server from database
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

	// Send a response back to the client
	response := map[string]string{"status": "ok", "verification": verf}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	// Email    string `json:"email"`
	Password string `json:"password"`
}

func (client *Client) registerHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a RegisterRequest struct
	fmt.Println("registerHandler")
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("json error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(req.Username, req.Password)
	// Validate the input
	if req.Username == "" {
		http.Error(w, "username cannot be empty", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "password cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if the username already exists in the database
	user_id, err := client.getUserByUsername(req.Username) // user_id =
	fmt.Println("Userid:", user_id)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user_id != 0 {
		http.Error(w, "username already exists", http.StatusBadRequest)
		return
	}

	// Insert the new user into the database
	err = client.insertUser(req.Username, req.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Added user into database")
	// Return a success response
	resp := map[string]bool{"ok": true}
	json.NewEncoder(w).Encode(resp)
}

func (client *Client) Run() {
	// Serve static files on website
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/register", client.registerHandler)
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

	// /*	Code for launch random available port for launching client GUI	*/
	// // Get a random available port
	// listener, err := net.Listen("tcp", "127.0.0.1:0")
	// if err != nil {
	// 	panic(err)
	// }
	// port := listener.Addr().(*net.TCPAddr).Port
	// portAddr := fmt.Sprintf(":%d", port)

	// // Start the server on launched port

	// log.Println("Starting server on port: ", port)
	// log.Printf("Visit Page on: http://localhost:%d/", port)
	// err = http.ListenAndServe(portAddr, nil)

	port := 9999
	log.Println("Starting client on port: ", port)
	log.Printf("Visit Page on: http://localhost:%d/", port)
	err = http.ListenAndServe(":9999", nil)

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

func (client *Client) chatHandler(w http.ResponseWriter, r *http.Request) {
	// Compile the chat template
	tmpl := template.Must(template.ParseFiles("./static/chat.html"))
	data := loadChatLogsFromDatabase(client)

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (client *Client) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as JSON
	fmt.Println("client.handle send message:")

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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

func (client *Client) UpdateName() bool {

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("client conn.write error: ", err)
		return false
	}

	return true

}
