package qserver

import (
	"connect-server/model"
	"connect-server/protocol"
	"fmt"
	"net"
	"sync"
)

type UDPServer struct {
	UserCache *model.UserCache
	IP        string
	Port      int
	Error     error
	Conn      *net.UDPConn
	Count     int
	Lock      sync.Mutex
	Sum       int
}

func NewUDPServer(ip string, port int, userCache *model.UserCache) *UDPServer {
	return &UDPServer{UserCache: userCache, IP: ip, Port: port, Count: 0, Sum: 0}
}

func (server *UDPServer) Add() {
	server.Lock.Lock()
	defer server.Lock.Unlock()
	server.Count += 1
}

func (server *UDPServer) Sub() {
	server.Lock.Lock()
	defer server.Lock.Unlock()
	server.Count -= 1
}

func (server *UDPServer) Start() {
	server.Conn, server.Error = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(server.IP), Port: server.Port})
	if server.Error != nil {
		fmt.Println("start udp listen failed")
		return
	}
	defer server.Conn.Close()
	server.Receive()
}

func (server *UDPServer) Receive() {
	data := make([]byte, 1500)
	for {
		number, remoteAddr, err := server.Conn.ReadFromUDP(data)
		if err != nil {
			continue
		}
		server.Sum += 1
		go server.Unpack(data[:number], remoteAddr)
	}
}

func (server *UDPServer) Unpack(data []byte, addr *net.UDPAddr) {
	if len(data) != 0 {
		Type := data[9]
		switch int8(Type) {
		case protocol.AUDIO:
			audio := protocol.AudioMsg{}
			err := audio.Decode(data[:])
			if err != nil {
				fmt.Printf("UDPServer Unpack Error:%v", err.Error())
				return
			}
			server.Audio(&audio, addr)
		default:
		}
	}
}

func (server *UDPServer) Audio(audio *protocol.AudioMsg, addr *net.UDPAddr) {
	server.Add()

	if fromUser := server.UserCache.Query(audio.FromUuid); fromUser == nil {
		fmt.Println("Not have fromUser ", audio.FromUuid)
		return
	}

	var destUser *model.User
	if destUser = server.UserCache.Query(audio.DestUuid); destUser == nil {
		fmt.Println("Not have destUser", audio.DestUuid)
		return
	}

	if destUser.Port == 0 {
		fmt.Println("The destUser's Port is 0")
		return
	}

	buffer, err := audio.Encode()
	if err != nil {
		fmt.Printf("UDPServer Audio Error: %v", err.Error())
		return
	}
	n, err := server.Conn.WriteToUDP(buffer, &net.UDPAddr{IP: destUser.IP, Port: int(destUser.Port)})
	if n == 0 || err != nil {
		fmt.Printf("UDPServer send data error: %v", err.Error())
		return
	}
}
