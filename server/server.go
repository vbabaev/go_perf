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

type packet struct {
	type_ uint32
	length uint32
	message string
}

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
		read := 0
		
		for read < 8 {
			sock.SetReadDeadline(time.Now().Add(60 * time.Second))
			rlen, _, _ := sock.ReadFromUDP(buf[read:])	
			read += rlen
		}

		length := read_int32(buf[0:4])
		type_ := read_int32(buf[4:8])
		for read < (int(length) + 8) {
			rlen, _, _ := sock.ReadFromUDP(buf[read:])	
			read += rlen	
		}
		message := string(buf[8:(8+length)])

		if type_ == 1 {
			save_cpu(message, db)
		}
		if type_ == 2 {
			save_mem(message, db)
		}

		if type_ == 3 {
			save_procs(message, db)
		}
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

func save_procs(raw string, db *sql.DB) {
	fields := strings.SplitN(raw, " ", 2)
	time, _ := strconv.ParseUint(fields[0], 10, 64)
	procs := strings.Split(fields[1], ";")
	for i := 0; i < len(procs); i++ {
		parts := strings.SplitN(procs[i], " ", 2)
		if (len(parts) == 2) {
			stm, err := db.Prepare("INSERT INTO proc (`time`, `proc`, `value`) VALUE(from_unixtime(?), ?, ?)")
			if err != nil {
				fmt.Println(err)
			}
			stm.Exec(time, parts[0], parts[1])
			stm.Close()	
		}
	}

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
