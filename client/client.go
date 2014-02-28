package client

import (
    "../stat"
    "../routine"
    "../report"
    "fmt"
    "net"
    "time"
    "container/list"
)

var (
    routines []*routine.Routine
)

func Run(host string, port int) {
    report_list := list.New()

    routines = make([]*routine.Routine, 0, 20)
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
}

func Stop() {
    for _, r := range routines {
        r.Stop()
    }
}
