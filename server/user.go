package main

import (
	"net"
	"strings"
)

type User struct {
	UserName string
	UserAddr string
	userChan chan string
	conn     net.Conn // every user has an unique connection to their client

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		UserName: userAddr,
		UserAddr: userAddr,
		userChan: make(chan string),
		conn:     conn,
		server:   server,
	}

	// Once new user built, set a go routine for it to listen message
	go user.ListenUserMessage()

	return user
}

// Listen user's channel, send message once received
func (user *User) ListenUserMessage() {
	for msg := range user.userChan {
		// msg := <-user.userChan
		// Only receive mssages from other users
		msgAddr := strings.Split(msg, "|")[0]
		if msgAddr != user.UserAddr {
			user.conn.Write([]byte(msg + "\n"))
		}
	}
	// No more message from user channel, so close its connection
	err := user.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (user *User) Online() {
	// New user added to online user map
	user.server.mapLock.Lock()
	user.server.OnlineUserMap[user.UserName] = user
	user.server.mapLock.Unlock()

	// Online anouncement for new user
	user.server.BroadCast(user, "ONLINE")

}

func (user *User) Offline() {
	// New user added to online user map
	user.server.mapLock.Lock()
	delete(user.server.OnlineUserMap, user.UserName)
	user.server.mapLock.Unlock()

	// Online anouncement for new user
	user.server.BroadCast(user, "OFFLINE")
}

func (user *User) SendMsg(msg string) {
	user.conn.Write(([]byte(msg)))
}

func (user *User) DoMessage(msg string) {
	// search online user command
	if msg == "who" {

		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineUserMap {
			onlineMsg := "[" + u.UserAddr + "]" + u.UserName + ": Online\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// rename|YOUR NEW NICK NAME
		newName := strings.Split(msg, "|")[1]

		// check if the new name is unique
		_, isExist := user.server.OnlineUserMap[newName]
		if isExist {
			user.SendMsg("This nickname already been used\n")
		} else {
			// delete the user from online map
			user.server.mapLock.Lock()
			delete(user.server.OnlineUserMap, user.UserName)
			user.server.OnlineUserMap[newName] = user
			user.server.mapLock.Unlock()

			user.UserName = newName
			// user.SendMsg("|system annoucement|You already changed your username to:" + user.UserName + "\n")
		}

	} else if len(msg) > 3 && msg[:3] == "to|" {
		to_username := strings.Split(msg, "|")[1]
		if to_username == "" {
			user.SendMsg("message format incorrect, follow the format: \"to|UserName|YourMessage\"\n")
		}

		// search the user in onlinemap
		to_user, isExist := user.server.OnlineUserMap[to_username]
		if !isExist {
			user.SendMsg("User not exist\n")
			return
		}

		// get the message and send it to the target user
		userMessage := strings.Split(msg, "|")[2]
		if userMessage == "" {
			user.SendMsg("Not message received, resend it\n")
		} else {
			to_user.SendMsg("(Private) " + user.UserName + ": " + userMessage)
		}

	} else {
		// chat to all user
		user.server.BroadCast(user, msg)
	}

}
