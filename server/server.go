package server

import (
    "../report"
    "database/sql"
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

func Run(host string, port int) {
    db, err := sql.Open("mysql", "root:123@/perf")
    if err != nil {
        return
    }
    db.SetMaxOpenConns(10)
    defer db.Close()
    var buf [4096]byte
    addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
    sock, _ := net.ListenUDP("udp", addr)
    header_size := report.GetHeaderSize()
    for {
        err = db.Ping()
        if err != nil {
            log.Fatalf("Error on opening database connection: %s", err.Error())
        }
        read := 0
        
        for uint32(read) < header_size {
            sock.SetReadDeadline(time.Now().Add(60 * time.Second))
            rlen, _, _ := sock.ReadFromUDP(buf[read:])
            read += rlen
        }

        time, length, type_ := report.ParseHeader(buf[0:header_size])

        for read < (int(length) + int(header_size)) {
            rlen, _, _ := sock.ReadFromUDP(buf[read:])
            read += rlen    
        }
        r := report.Load(time, type_, buf[header_size:(header_size+length)])

        if type_ == report.TYPE_CPU {
            save_cpu(r, db)
        }
        if type_ == report.TYPE_MEM {
            save_mem(r, db)
        }
        if type_ == report.TYPE_PROC {
            save_procs(r, db)
        }
        if type_ == report.TYPE_IO {
            save_io(r, db)
        }
    }
}

func Stop() {
    
}

func save_cpu(r report.Report, db *sql.DB) {
    cpu, _ := strconv.ParseFloat(r.Text(), 64)
    stm, err := db.Prepare("INSERT INTO cpu (`time`, `value`) VALUE(from_unixtime(?), ?)")
    if err != nil {
        fmt.Println(err)
        return
    }
    stm.Exec(r.Time(), cpu)
    defer stm.Close()
}

func save_io(r report.Report, db *sql.DB) {
    io, _ := strconv.ParseInt(r.Text(), 10, 64)
    stm, err := db.Prepare("INSERT INTO io (`time`, `value`) VALUE(from_unixtime(?), ?)")
    if err != nil {
        fmt.Println(err)
        return
    }
    stm.Exec(r.Time(), io)
    defer stm.Close()
}

func save_procs(r report.Report, db *sql.DB) {
    procs := strings.Split(r.Text(), ";")
    stm, err := db.Prepare("INSERT INTO proc (`time`, `proc`, `value`) VALUE(from_unixtime(?), ?, ?)")
    if err != nil {
        fmt.Println(err)
        return
    }
    for i := 0; i < len(procs); i++ {
        parts := strings.SplitN(procs[i], " ", 2)
        if (len(parts) == 2) {
            stm.Exec(r.Time(), parts[0], parts[1])
        }
    }
    stm.Close()    

}

func save_mem(r report.Report, db *sql.DB) {
    fields := strings.Fields(r.Text())
    mem, _ := strconv.ParseFloat(fields[0], 64)
    swap, _ := strconv.ParseFloat(fields[1], 64)

    stm, err := db.Prepare("INSERT INTO mem (`time`, `value`) VALUE(from_unixtime(?), ?)")
    if err != nil {
        fmt.Println(err)
    }
    stm.Exec(r.Time(), mem)
    stm.Close()

    stm, err = db.Prepare("INSERT INTO swap (`time`, `value`) VALUE(from_unixtime(?), ?)")
    if err != nil {
        fmt.Println(err)
    }
    stm.Exec(r.Time(), swap)
    stm.Close()
}
