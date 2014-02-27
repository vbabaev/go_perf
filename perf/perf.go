package main

import (
	"../stat"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var routinesNextId = 0

type routineFunc func()
type t_routine struct {
	id     int
	pause  bool
	status bool
}

type report_message struct {
	size  uint32
	type_ uint32
	text  string
}

type link_list_node struct {
	item report_message
	next *link_list_node
}

type linked_list struct {
	tail *link_list_node
	head *link_list_node
}

var tail link_list_node = link_list_node{report_message{0, 0, ""}, nil}
var head link_list_node = link_list_node{report_message{0, 0, ""}, &tail}
var report_list linked_list = linked_list{&tail, &head}

func linked_list_push(report report_message) {
	var new_tail link_list_node = link_list_node{report_message{0, 0, ""}, nil}
	report_list.tail.item = report
	report_list.tail.next = &new_tail
	report_list.tail = &new_tail
}

func linked_list_pop() (report report_message) {
	if report_list.head.next == report_list.tail {
		return
	}
	report = report_list.head.next.item
	report_list.head = report_list.head.next
	return
}

func create_report_message(type_ uint32, message string) report_message {
	return report_message{uint32(len(message)), type_, message}
}

func main() {
	var host string
	var port int
	flag.StringVar(&host, "h", "localhost", "server's hostname")
	flag.IntVar(&port, "p", 30000, "server's port name")
	flag.Parse()

	routines := make([]*t_routine, 0, 20)
	getCpuUsage := stat.GetCPUsageProvider()
	time.Sleep(1 * time.Second)

	cpu := func() {
		report := create_report_message(1, fmt.Sprintf("%s %.2f", GetCurrentTimestamp(), getCpuUsage()))
		linked_list_push(report)
	}

	mem := func() {
		total, free, swap_total, swap_free, cached := stat.GetMem()
		pmem := float64(total-free-cached) / float64(total)
		pswap := float64(swap_total-swap_free) / float64(swap_total)
		report := create_report_message(2, fmt.Sprintf("%s %.2f %.2f", GetCurrentTimestamp(), pmem, pswap))
		linked_list_push(report)
	}

	procs := func() {
		p_procs := stat.GetTopProcs(5)
		var message string = ""
		message += fmt.Sprintf("%s ", GetCurrentTimestamp())
		for _, proc := range p_procs {
			message += fmt.Sprintf("%s %.2f;", proc.Name, proc.Percentage)
		}
		report := create_report_message(3, message)
		linked_list_push(report)

	}

	send := func() {
		serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			fmt.Println(err)
			return
		}
		con, err := net.DialUDP("udp", nil, serverAddr)

		buf := make([]byte, 1024)
		for pointer := report_list.head.next; pointer != report_list.tail; pointer = pointer.next {
			report := linked_list_pop()
			binary.LittleEndian.PutUint32(buf[0:4], uint32(report.size))
			binary.LittleEndian.PutUint32(buf[4:8], uint32(report.type_))
			copy(buf[8:], report.text)
			send_len := 8 + len(report.text)
			con.Write(buf[0:send_len])
		}
		con.Close()
	}

	routines = append(routines, runRoutineWithPeriod(cpu, 5 * time.Second))
	routines = append(routines, runRoutineWithPeriod(mem, 5 * time.Second))
	routines = append(routines, runRoutineWithPeriod(procs, 1 * time.Second))
	routines = append(routines, runRoutineWithPeriod(send, 1 * time.Second))

	var status bool = true

	for status {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			status = false
		case syscall.SIGHUP:
			status = false
		case syscall.SIGTERM:
			status = false
		}
	}
	for _, r := range routines {
		r.status = false
	}
}

func runRoutineWithPeriod(routine routineFunc, period time.Duration) *t_routine {
	var this_routine t_routine
	this_routine.id = getRoutineId()
	this_routine.pause = true
	this_routine.status = true
	go func() {
		for {
			if this_routine.status == false {
				break
			}
			start := time.Now()
			if this_routine.pause == true {
				routine()
			}
			delta := time.Since(start)
			if delta < period {
				time.Sleep(period - delta)
			}
		}
	}()

	return &this_routine
}

func getRoutineId() (result int) {
	result = routinesNextId
	routinesNextId++
	return
}

func GetCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
