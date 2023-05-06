package main

import "flag"

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Server IP")
	flag.IntVar(&serverPort, "port", 8888, "Server Port")
}
func main() {
	flag.Parse()
	server := NewServer(serverIp, serverPort)
	server.Start()
}
