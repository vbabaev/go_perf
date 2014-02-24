package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

func read_int32(data []byte) (ret int32) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

var host string
var port int

func main() {
	flag.StringVar(&host, "h", "localhost", "server's hostname")
	flag.IntVar(&port, "p", 30000, "server's port name")
	flag.Parse()

	db, err := sql.Open("mysql", "root:123@/perf")
	if err != nil {
		return
	}
	db.SetMaxOpenConns(10)
	defer db.Close()
	var buf [4096]byte
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	sock, _ := net.ListenUDP("udp", addr)
	for {
		err = db.Ping()
		if err != nil {
			log.Fatalf("Error on opening database connection: %s", err.Error())
		}

		tx, _ := db.Begin()
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
		if type_ == 1 {
			save_cpu(message, db)
		}
		if type_ == 2 {
			save_mem(message, db)
		}
		tx.Commit()
	}
}

func save_cpu(raw string, db *sql.DB) {
	fields := strings.Fields(raw)
	time, _ := strconv.ParseUint(fields[0], 10, 64)
	cpu, _ := strconv.ParseFloat(fields[1], 64)
	stm, err := db.Prepare("INSERT INTO cpu (`time`, `value`) VALUE(from_unixtime(?), ?)")
	if err != nil {
		fmt.Println(err)
	}
	stm.Exec(time, cpu)
	defer stm.Close()
}

func save_mem(raw string, db *sql.DB) {
	fields := strings.Fields(raw)
	time, _ := strconv.ParseUint(fields[0], 10, 64)
	mem, _ := strconv.ParseFloat(fields[1], 64)
	swap, _ := strconv.ParseFloat(fields[2], 64)
	stm, err := db.Prepare("INSERT INTO mem (`time`, `value`) VALUE(from_unixtime(?), ?)")
	if err != nil {
		fmt.Println(err)
	}
	stm.Exec(time, mem)
	stm.Close()

	stm, err = db.Prepare("INSERT INTO swap (`time`, `value`) VALUE(from_unixtime(?), ?)")
	if err != nil {
		fmt.Println(err)
	}
	stm.Exec(time, swap)
	stm.Close()
}
