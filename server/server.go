package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip         string
	Port       int
	serverChan chan string

	// mapLock and online user map
	OnlineUserMap map[string]*User
	mapLock       sync.RWMutex
}

// Server interface
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]*User),
		serverChan:    make(chan string),
	}

	return server
}

func (server *Server) BroadCast(user *User, msg string) {
	// Broadcasting new online user
	announce := user.UserAddr + "|" + user.UserName + "|" + msg
	log.Printf("Broadcast from %s", user.UserAddr)
	server.serverChan <- announce

}

func (server *Server) ListenServerChannel() {
	// Once new message got in server channel, anounce it to all online users

	for {
		msg := <-server.serverChan

		server.mapLock.Lock()
		for _, userCli := range server.OnlineUserMap {
			userCli.userChan <- msg
		}
		server.mapLock.Unlock()

	}

}

func (server *Server) Handler(conn net.Conn) {
	fmt.Println("Connection built:", conn)

	// Once a connection built, assign a user class for this
	user := NewUser(conn, server)

	user.Online()

	// Channel checks whether user is active (Otherwise kick this user)
	isActive := make(chan bool)

	// Handle message sent from client
	go func() {
		buff := make([]byte, 4096) // message buffer
		for {
			len_message, err := conn.Read(buff)
			if len_message == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn read error:", err)
				return
			}

			// save message withought the last '\n'
			msg := string(buff[:len_message-1])

			// handle with user message
			user.DoMessage(msg)

			// Handling message means user is active
			isActive <- true

		}
	}()

	for {
		select {
		case <-isActive:
			// user is active, reset the timer
			// pass
		case <-time.After(time.Second * 1800):
			// Not active over time, kick this user
			// user.SendMsg("You're disconected since long time inactive (5 mins)")

			// delete user
			close(user.userChan)
			// conn.Close()

			// exit from this handler
			runtime.Goexit()
		}
	}
}

func (server *Server) Start() {
	// socket listening
	Listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("socket Listen error:", err)
		return
	}
	fmt.Println("Server launched on ", server.Ip, ":", server.Port)
	defer Listener.Close()

	for {
		conn, err := Listener.Accept() // net.Dial
		if err != nil {
			fmt.Println("Listener accept error", err)
			continue
		}

		// parrallel
		go server.ListenServerChannel()

		go server.Handler(conn)
	}

}
