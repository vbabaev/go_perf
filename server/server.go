package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func read_int32(data []byte) (ret int32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func main() {
	var buf [4096]byte
	// addr, _ := net.ResolveUDPAddr("udp", "146.185.149.162:30003")
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:30003")
	sock, _ := net.ListenUDP("udp", addr)
	for {
		readBytes := 0
		for readBytes < 8 {
			sock.SetReadDeadline(time.Now().Add(5 * time.Second))
			rlen, _, _ := sock.ReadFromUDP(buf[readBytes:])
			readBytes += rlen
		}
		length := read_int32(buf[0:4])
		type_ := read_int32(buf[4:8])

		readBytes = 0
		var message string = ""
		for int32(readBytes) < length {
			sock.SetReadDeadline(time.Now().Add(5 * time.Second))
			rlen, _, _ := sock.ReadFromUDP(buf[readBytes:])
			message += string(buf[:rlen])
			readBytes += rlen
		}
		if type_ == 0 {

		}
		fmt.Println(message)
	}
}
