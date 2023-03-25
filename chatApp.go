package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// var upgrader = websocket.Upgrader{}

func main() {
	// Serve the static files for the frontend
	http.Handle("/", http.FileServer(http.Dir("static")))

	// template
	http.HandleFunc("/chat", chatHandler)

	// WebSocket endpoint for the chat messages
	http.HandleFunc("/send-message", handleSendMessage)

	// Start the server on port 8080
	log.Println("Starting server on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
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

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Parse the request body as JSON
	fmt.Println("handlesend message")
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the value of the input_message field from the request body
	inputMessage, ok := data["input_message"]
	if !ok {
		http.Error(w, "Missing input_message field in request body", http.StatusBadRequest)
		return
	}

	// Do something with the input message, such as send it to other users in the chat
	fmt.Println(inputMessage)
	// data["retrieved_message"] = "~frontend"
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
