package stat

import (
    "fmt"
    "io/ioutil"
    "strconv"
    "strings"
    "os/exec"
    "bytes"
    "sort"
)

type Proc struct {
    Name string
    Percentage float32
}

type ByPerc []Proc

func (a ByPerc) Len() int           { return len(a) }
func (a ByPerc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPerc) Less(i, j int) bool { return a[i].Percentage > a[j].Percentage }

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

func GetTopProcs(count int) (s_procs []Proc) {
    cmd := exec.Command("/bin/ps", "hx", "-ocomm,pcpu", "--sort=pcpu")
    cmd.Stdin = strings.NewReader("some input")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return  
    } 
    contents := out.String()
    lines := strings.Split(string(contents), "\n")

    proc_list := make(map[string]float64)

    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) == 2 {
            name := fields[0]
            percentage, _ := strconv.ParseFloat(fields[1], 64)
            if _,ok := proc_list[name]; !ok && percentage != 0 {
                proc_list[name] = 0
            } 
            if percentage != 0 {
                proc_list[name] += percentage
            }
            
        }
    }

    s_procs = make([]Proc, len(proc_list), len(proc_list))
    i := 0
    for name, percentage := range proc_list {
        s_procs[i] = Proc{name, float32(percentage)}
        i++
    }
    sort.Sort(ByPerc(s_procs))
    if len(s_procs) < count {
        count = len(s_procs)
    }

    s_procs = s_procs[0:count]
    return 
}
