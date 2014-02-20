package stat

import (
    "fmt"
    "io/ioutil"
    "strconv"
    "strings"
)

func GetCPUsageProvider() (func() float64) {
    idle0, total0 := GetCPU()
    return func() float64 {
        idle1, total1 := GetCPU() 
        idleTicks := float64(idle1 - idle0)
        totalTicks := float64(total1 - total0)
        cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
        idle0 = idle1
        total0 = total1
        return cpuUsage
    }
}

func GetCPU() (idle, total uint64) {
    contents, err := ioutil.ReadFile("/proc/stat")
    if err != nil {
        return
    }
    lines := strings.Split(string(contents), "\n")
    for _, line := range(lines) {
        fields := strings.Fields(line)
        if fields[0] == "cpu" {
            numFields := len(fields)
            for i := 1; i < numFields; i++ {
                val, err := strconv.ParseUint(fields[i], 10, 64)
                if err != nil {
                    fmt.Println("Error: ", i, fields[i], err)
                }
                total += val
                if i == 4 {
                    idle = val
                }
            }
            return
        }
    }
    return
}

func GetMem() (total, free, swap_total, swap_free, cached uint64) {
    contents, err := ioutil.ReadFile("/proc/meminfo")
    if err != nil {
        return
    }
    lines := strings.Split(string(contents), "\n")
    for _, line := range(lines) {
        fields := strings.Fields(line)
        if len(fields) > 0 && fields[0] == "MemTotal:" {
            total, _ = strconv.ParseUint(fields[1], 10, 64)
        }
        
        if len(fields) > 0 && fields[0] == "MemFree:" {
            free, _ = strconv.ParseUint(fields[1], 10, 64)
        }

        if len(fields) > 0 && fields[0] == "SwapTotal:" {
            swap_total, _ = strconv.ParseUint(fields[1], 10, 64)
        }
        
        if len(fields) > 0 && fields[0] == "SwapFree:" {
            swap_free, _ = strconv.ParseUint(fields[1], 10, 64)
        }

        if len(fields) > 0 && fields[0] == "Cached:" {
            cached, _ = strconv.ParseUint(fields[1], 10, 64)
        }
    }
    return
}
