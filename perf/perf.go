package main

import (
	"../stat"
	"../routine"
	"../report"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type link_list_node struct {
	item *report.Report
	next *link_list_node
}

type linked_list struct {
	tail *link_list_node
	head *link_list_node
}

var tail link_list_node = link_list_node{nil, nil}
var head link_list_node = link_list_node{nil, &tail}

var report_list linked_list = linked_list{&tail, &head}

func linked_list_push(report report.Report) {
	next_tail := link_list_node{nil, nil} 
	report_list.tail.next = &next_tail
	report_list.tail.item = &report
	report_list.tail = &next_tail
}

func linked_list_pop() (report *report.Report) {
	if report_list.head.next == report_list.tail {
		return
	}
	report = report_list.head.next.item
	report_list.head = report_list.head.next
	return
}

func main() {
	var host string
	var port int
	flag.StringVar(&host, "h", "localhost", "server's hostname")
	flag.IntVar(&port, "p", 30000, "server's port name")
	flag.Parse()

	routines := make([]*routine.Routine, 0, 20)
	getCpuUsage := stat.GetCPUsageProvider()
	time.Sleep(1 * time.Second)

	cpu := func() {
		report := report.Create(report.TYPE_CPU, fmt.Sprintf("%.3f", getCpuUsage()))
		linked_list_push(report)
	}

	mem := func() {
		total, free, swap_total, swap_free, cached := stat.GetMem()
		pmem := float64(total-free-cached) / float64(total)
		pswap := float64(swap_total-swap_free) / float64(swap_total)
		report := report.Create(report.TYPE_MEM, fmt.Sprintf("%.3f %.3f", pmem, pswap))
		linked_list_push(report)
	}

	procs := func() {
		p_procs := stat.GetTopProcs(5)
		var message string = ""
		for _, proc := range p_procs {
			message += fmt.Sprintf("%s %.2f;", proc.Name, proc.Percentage)
		}
		report := report.Create(report.TYPE_PROC, message)
		linked_list_push(report)
	}

	send := func() {
		serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

		if err != nil {
			fmt.Println(err)
			return
		}
		con, err := net.DialUDP("udp", nil, serverAddr)
		
		for pointer := report_list.head.next; pointer != report_list.tail; pointer = pointer.next {
			con.Write(pointer.item.Pack())
			linked_list_pop()
		}
		con.Close()
	}

	routines = append(routines, routine.Create(cpu, time.Second))
	routines = append(routines, routine.Create(mem, time.Second))
	routines = append(routines, routine.Create(procs, 5 * time.Second))
	routines = append(routines, routine.Create(send, 1 * time.Second))

	for _, routine := range routines {
		routine.Start()
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
	for _, r := range routines {
		r.Stop()
	}
}
