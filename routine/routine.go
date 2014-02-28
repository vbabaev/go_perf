package routine

import (
    "time"
)

var next_routine_id uint32 = 0

type (
    routine_method func()
    Routine struct {
        id uint32
        paused bool
        status bool
        duration time.Duration
        method routine_method
    }
)

func Create(method routine_method, duration time.Duration) *Routine {
    next_routine_id++
    return &Routine{next_routine_id, false, false, duration, method}
}

func (r * Routine) process() {
    r.method()
}

func (r * Routine) cycle() {
    for r.status {
        start := time.Now()
        if !r.paused {
            r.process()
        }
        delta := time.Since(start)
        if delta < r.duration {
            time.Sleep(r.duration - delta)
        }
    }
}

func (r * Routine) Start() {
    if !r.status {
        r.status = true
        go r.cycle()
    }
}

func (r * Routine) Stop() {
    r.status = false
}

func (r * Routine) Pause() {
    r.paused = true
}

func (r * Routine) Unpause() {
    r.paused = false
}

func (r * Routine) Id() uint32 {
    return r.id
}