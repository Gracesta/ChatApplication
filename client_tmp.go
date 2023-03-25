package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Client struct {
	Ip   string
	Port int
	Name string
	flag int

	conn net.Conn
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

func (client *Client) Run() {
	// for client.flag != 0 {
	// 	for client.menu() != true {
	// 		// display menu while input illegal
	// 	}

	// 	switch client.flag {
	// 	case 1:
	// 		client.PublicChat()
	// 		break
	// 	case 2:
	// 		client.PrivateChat()
	// 		break
	// 	case 3:
	// 		client.UpdateName()
	// 		break
	// 	}
	// }
	// Web browser to listen to user's input
	// Serve the static files for the frontend
	go http.Handle("/", http.FileServer(http.Dir("static")))

	// template
	go http.HandleFunc("/chat", client.chatHandler)

	// WebSocket endpoint for the chat messages
	go http.HandleFunc("/send-message", client.handleSendMessage)

	// // WebSocket endpoint for the chat messages
	// go http.HandleFunc("/receive-message", client.handleReceivedMessage)

	// Register the SSE endpoint
	// http.HandleFunc("/message-stream", messageStreamHandler)

	// Start the server on port 8080
	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// for {

	// }
}

type Message struct {
	InputMessage string `json:"input_message"`
}

func messageStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a channel to send messages to the client
	messageChan := make(chan Message)

	// Start a goroutine to send messages to the client
	go func() {
		for {
			// Wait for a message from the channel
			message := <-messageChan
			fmt.Println("message channel:", message)
			// Marshal the message as JSON
			data, err := json.Marshal(message)
			if err != nil {
				fmt.Println("Error marshaling message:", err)
				continue
			}

			// Send the message to the client
			fmt.Fprintf(w, "data: %s\n\n", data)

			// Flush the response to ensure that the message is sent immediately
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}()

	// Loop indefinitely to simulate sending messages from the backend
	for i := 0; ; i++ {
		// Create a message with the current time
		message := Message{
			InputMessage: fmt.Sprintf("Message %d sent at %s", i, time.Now().Format("15:04:05")),
		}

		// Send the message to the channel
		messageChan <- message

		// Sleep for 2 seconds before sending the next message
		time.Sleep(1 * time.Second)
	}
}

func (client *Client) HandleResponse() {
	// handle the message from server
	// // Once read message from conn, display it
	// io.Copy(os.Stdout, client.conn)
	buf := make([]byte, 1024)
	for {
		n, err := client.conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		msg := string(buf[:n])
		fmt.Println("Received data from client.conn:", msg)

		// data := map[string]string{
		// 	"input_message": msg,
		// }

		// Convert the data to JSON
		// jsonData, err := json.Marshal(data)
		// if err != nil {
		// 	panic(err)
		// }

		// Send the data to the frontend in a separate goroutine
		// go func() {
		// 	// Make the HTTP POST request
		// 	resp, err := http.Post("http://localhost:8080/receive-message", "application/json", bytes.NewBuffer(jsonData))
		// 	if err != nil {
		// 		fmt.Println("error on sending received message to client browser")
		// 		panic(err)
		// 	}
		// 	// fmt.Println("response to http", resp)
		// 	defer resp.Body.Close()
		// }()

	}

}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Server IP")
	flag.IntVar(&serverPort, "port", 8888, "Server Port")
}

// var upgrader = websocket.Upgrader{}

func main() {
	// command line parse
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> Link to Server Falied <<<<<")
		return
	}

	fmt.Println(">>>>> Link to Server Succedded <<<<<")

	// if client.flag > 0 && client.flag < 4 {
	// 	client.menu()
	// } else {
	// 	fmt.Println("Input number within range")
	// }

	// go routin for handling message from server
	go client.HandleResponse()

	client.Run()

}

/*
----------------------   Web Browser -----------------------------------------------------
*/
func (client *Client) chatHandler(w http.ResponseWriter, r *http.Request) {
	// Compile the chat template
	fmt.Println("chatHandler")
	tmpl := template.Must(template.ParseFiles("./static/chat.html"))

	// Render the template with any necessary data
	data := struct {
		Title string
	}{
		Title: "Chat Application",
	}
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
	fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("json data", data)
	// Get the value of the input_message field from the request body
	chatMsg, ok := data["input_message"]
	if !ok {
		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
		return
	}

	// Do something with the input message, such as send it to other users in the chat
	fmt.Println("Broadcast message to other users:", chatMsg)
	if len(chatMsg) != 0 {
		sendMsg := []byte(chatMsg)
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("client conn.write error: ", err)
		}

	}

	message_backend := "~backend"
	// Send a response back to the client
	response := map[string]string{"status": "ok", "message_backend": message_backend}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// func (client *Client) handleReceivedMessage(w http.ResponseWriter, r *http.Request) {

// 	var data map[string]string
// 	err := json.NewDecoder(r.Body).Decode(&data)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// Get the value of the input_message field from the request body
// 	chatMsg, ok := data["input_message"]
// 	if !ok {
// 		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Do something with the input message, such as send it to other users in the chat
// 	fmt.Println("Received message from other users:", chatMsg)
// 	// if len(chatMsg) != 0 {
// 	// 	sendMsg := []byte(chatMsg)
// 	// 	_, err := client.conn.Write([]byte(sendMsg))
// 	// 	if err != nil {
// 	// 		fmt.Println("client conn.write error: ", err)
// 	// 	}

// 	// }

// 	// message_backend := "~backend"
// 	// Send a response back to the client
// 	response := map[string]string{"status": "ok", "message_backend": chatMsg}
// 	w.Header().Set("Content-Type", "application/json")
// 	fmt.Println(response)
// 	err = json.NewEncoder(w).Encode(response)
// 	if err != nil {
// 		fmt.Println("error")
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

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

func (client *Client) PublicChat() {
	var chatMsg string
	var errBufio error
	in := bufio.NewReader(os.Stdin)

	fmt.Println(">>>>> Input the message you want to send (input \"exit\" to exit)")
	chatMsg, errBufio = in.ReadString('\n')
	if errBufio != nil {
		fmt.Println("reading string error", errBufio)
	}
	// fmt.Scanln(&chatMsg)

	for chatMsg != "exit\r\n" {

		// send message to usre conn to broadcast
		if len(chatMsg) != 0 {
			sendMsg := chatMsg
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("client conn.write error: ", err)
				break
			}

		}

		chatMsg = ""
		fmt.Println(">>>>> Input the message you want to send (input \"exit\" to exit)")
		chatMsg, errBufio = in.ReadString('\n')
		if errBufio != nil {
			fmt.Println("reading string error", errBufio)
		}
		// fmt.Scanln(&chatMsg)
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
