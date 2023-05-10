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
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

var (
	OnlineClientrMap map[int]*Client
	clientMutex      sync.Mutex
)
var mux = http.NewServeMux()

// var msgChan chan map[string]string
var client_db *sql.DB

// var wsConn *websocket.Conn

// key: request remote address, value: channel to send data
var reqToChanMap map[string](chan map[string]string) // mapLock      sync.RWMutex

type Client struct {
	Ip   string
	Port int
	Name string

	conn    net.Conn        // connection for server
	wsConn  *websocket.Conn // connection for websocket
	db      *sql.DB
	user_id int

	// msgChan *chan map[string]string // meesage channel for message from front-end (AJAX)
}

func NewClient(ip string, port int) *Client {
	client := &Client{
		Ip:   ip,
		Port: port,
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

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to WebSocket
	fmt.Println("handlewebsocket handler from:", r.RemoteAddr)
	fmt.Println(r.URL)
	wsConn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	params := r.URL.Query()
	fmt.Println(params)
	user_id_string := params.Get("userId")
	fmt.Println("Connection from new user to websocket BUILT, user_id:", user_id_string)
	user_id, err := strconv.Atoi(user_id_string)
	if err != nil {
		log.Fatal("error in converting string to int in handlwebsocket")
	}
	clientMutex.Lock()
	OnlineClientrMap[user_id].wsConn = wsConn
	clientMutex.Unlock()

	fmt.Println(OnlineClientrMap[user_id])
	for {

	}
}
func loginVerificationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("loginHandler")
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get the value of the input_message field from the request body

	username, ok := data["username"]
	password, ok := data["password"]
	fmt.Println(username, password)
	if !ok {
		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
		return
	}
	var user_id int

	// checkLogin function will return 'user_id' and 'error'
	user_id, err = checkLogin(client_db, username, password)
	if err != nil {
		fmt.Println("Failed to login with database: %v", err)
	}

	// Verification on login process
	var verf string

	if user_id != 0 {
		// Check if this user already loggined
		preClient, exists := OnlineClientrMap[user_id]
		if exists {
			preClient.conn.Close()
			preClient.wsConn.Close()
		}

		client := NewClient(serverIp, serverPort)
		fmt.Println(">>>>> Link to Server Succedded <<<<<")

		client.user_id = user_id
		client.Name = username
		client.UpdateName() // retrieve username to server from database

		clientMutex.Lock()
		OnlineClientrMap[user_id] = client
		clientMutex.Unlock()

		fmt.Println("Login succeeded, user_id is:", user_id)
		verf = "TRUE"
		go client.Run()

	} else {
		verf = "FALSE"
	}

	// Send a response back to the client
	response := map[string]string{"status": "ok", "verification": verf, "user_id": strconv.Itoa(user_id)}
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
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
	user_id, err := getUserByUsername(req.Username, client_db) // user_id =
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
	err = insertUser(req.Username, req.Password, client_db)
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

func (client *Client) RetrieveMessage(msgByte []byte) {
	if err := client.wsConn.WriteMessage(websocket.TextMessage, msgByte); err != nil {
		fmt.Println("wsconn write error", err)
	}
}

func getJSTTimeStamp(timestamp string) string {
	// Parse the timestamp as a time.Time object
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		fmt.Println("Failed to parse timestamp:", err)
		return timestamp
	}

	// Load the JST time zone
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println("Failed to load JST time zone:", err)
		return timestamp
	}

	// Convert the timestamp to JST
	jstTime := t.In(jst)

	// Format the JST time as a string
	jstTimestamp := jstTime.Format(time.RFC3339)
	return jstTimestamp

}

func (client *Client) SendMessage(data map[string]string) {
	fmt.Println("get message from msgChan channel")
	// new message fron front-end
	// Get the value of the input_message field from the request body
	chatMsg, ok := data["input_message"]
	timestamp, ok := data["timestamp"]
	timestamp = getJSTTimeStamp(timestamp)
	fmt.Println(timestamp)
	if !ok {
		log.Fatal("Missing input_message field in request body", http.StatusBadRequest)
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
	stmt, err := client_db.Prepare("INSERT INTO chat_logs(user_id, message, timestamp) VALUES(?, ?, ?)")
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

}

func (client *Client) Run() {
	fmt.Println("Client Running")

	go func() {
		for {
			// Read data from client.conn
			buf := make([]byte, 1024)
			n, err := client.conn.Read(buf)
			if err != nil {
				break
			}
			msg := string(buf[:n])
			fmt.Println("Receive Message in websocket from user conn and try to write it to wsConn:", msg)

			// Send data to frontend over WebSocket
			client.RetrieveMessage(buf[:n])
			// if err := client.wsConn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			// 	fmt.Println("wsconn write error", err)
			// 	break
			// }
		}
	}()
}

var serverIp string
var serverPort int
var clientPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Server IP")
	flag.IntVar(&serverPort, "port", 8888, "Server Port")
	flag.IntVar(&clientPort, "client_port", 9999, "Client Port")
}

func main() {
	// command line parse
	flag.Parse()

	OnlineClientrMap = make(map[int]*Client)
	go launchDatabase()
	// Here for login process

	go mux.Handle("/", http.FileServer(http.Dir("static")))
	go mux.HandleFunc("/ws", handleWebSocket)
	go mux.HandleFunc("/register", registerHandler)
	go mux.HandleFunc("/chat", chatHandler)
	go mux.HandleFunc("/login-verif", loginVerificationHandler)
	go mux.HandleFunc("/send-message", handleSendMessage)

	port := clientPort // clientPort initialized by commandline arugments
	log.Println("Starting client on port: ", port)
	log.Printf("Visit Page on: http://localhost:%d/", port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	// Start server
	go log.Fatal(server.ListenAndServe())

}

func launchDatabase() {
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

	client_db = db

	log.Println("Database launched", db)
	for {

	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Compile the chat template
	tmpl := template.Must(template.ParseFiles("./static/chat.html"))
	data := loadChatLogsFromDatabase(client_db)

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as JSON
	fmt.Println("handle send message from:", r.RemoteAddr)
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	//data fields: input_message, user_id, timestamp
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user_id, err := strconv.Atoi(data["user_id"])
	if err != nil {
		log.Fatal("string convertion error in handle send message")
	}

	// send data to taget client
	target_client, exists := OnlineClientrMap[user_id]
	if !exists {
		log.Printf("Invalid send message from inactive user")
		return
	}
	target_client.SendMessage(data)

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
