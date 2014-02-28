package main

import (
	"../client"
	"../server"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var host string
	var port int
	var is_server bool
	flag.StringVar(&host, "h", "localhost", "server's hostname")
	flag.IntVar(&port, "p", 30000, "server's port name")
	flag.BoolVar(&is_server, "S", false, "server type")
	flag.Parse()

	if is_server {
		server.Run(host, port)
	} else {
		client.Run(host, port)
	}
	
    status := true
    for status {
        signalChannel := make(chan os.Signal, 2)
        signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
        sig := <-signalChannel
        switch sig {
        case os.Interrupt:
            status = false
            break
        case syscall.SIGHUP:
            status = false
            break
        case syscall.SIGTERM:
            status = false
            break
        }
    }

	if is_server {
		server.Stop()
	} else {
		client.Stop()
	}
}
