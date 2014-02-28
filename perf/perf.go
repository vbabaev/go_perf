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
	"container/list"
)

func main() {
	var host string
	var port int
	flag.StringVar(&host, "h", "localhost", "server's hostname")
	flag.IntVar(&port, "p", 30000, "server's port name")
	flag.Parse()

	report_list := list.New()

	routines := make([]*routine.Routine, 0, 20)
	getCpuUsage := stat.GetCPUsageProvider()
	time.Sleep(1 * time.Second)

	cpu := func() {
		report := report.Create(report.TYPE_CPU, fmt.Sprintf("%.3f", getCpuUsage()))
		report_list.PushBack(report)
	}

	mem := func() {
		total, free, swap_total, swap_free, cached := stat.GetMem()
		pmem := float64(total-free-cached) / float64(total)
		pswap := float64(swap_total-swap_free) / float64(swap_total)
		report := report.Create(report.TYPE_MEM, fmt.Sprintf("%.3f %.3f", pmem, pswap))
		report_list.PushBack(report)
	}

	procs := func() {
		p_procs := stat.GetTopProcs(5)
		var message string = ""
		for _, proc := range p_procs {
			message += fmt.Sprintf("%s %.2f;", proc.Name, proc.Percentage)
		}
		report := report.Create(report.TYPE_PROC, message)
		report_list.PushBack(report)
	}

	send := func() {
		serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))

		if err != nil {
			fmt.Println(err)
			return
		}
		con, err := net.DialUDP("udp", nil, serverAddr)

		for e := report_list.Front(); e != nil; e = report_list.Front() {
			report := e.Value.(report.Report)
			con.Write(report.Pack())
			report_list.Remove(e)
		}	
		
		con.Close()
	}

	routines = append(routines, routine.Create(cpu, time.Second))
	routines = append(routines, routine.Create(mem, time.Second))
	routines = append(routines, routine.Create(procs, 5 * time.Second))
	routines = append(routines, routine.Create(send, 45 * time.Second))

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
