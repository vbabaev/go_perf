package main

import (
    "fmt"
    "time"
    "github.com/vbabaev/stat"
)

var routinesNextId = 0

type routineFunc func()
type t_routine struct {
    id int
    pause bool
    status bool
}

func main() {
    routines := make([]*t_routine, 0, 20)
    getCpuUsage := stat.GetCPUsageProvider()
    time.Sleep(1 * time.Second)

    cpu := func () {
        fmt.Printf("%s %f\n", GetCurrentTimestamp(), getCpuUsage())
    }   
    

    mem := func () {
        total, free, swap_total, swap_free, cached := stat.GetMem()
        pmem := float64(total - free - cached) / float64(total)
        pswap := float64(swap_total - swap_free) / float64(swap_total)
        fmt.Printf("%s %f %f\n", GetCurrentTimestamp(), pmem, pswap)
    }   

    routines = append(routines, runRoutineWithPeriod(cpu, 1 * time.Second))
    routines = append(routines, runRoutineWithPeriod(mem, 1 * time.Second))
    
    time.Sleep(2 * time.Second)
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
            if (delta < period) {
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
    t := time.Now()
    return fmt.Sprintf("%.4d-%.2d-%.2d %.2d-%.2d-%.2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}
