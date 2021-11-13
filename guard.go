package procguard

import (
	"fmt"
	"path/filepath"
	"time"
)

type (
	logger   func(int, string, ...interface{})
	Guardian struct {
		Slots slots `json:"slots"` //时段定义
		Procs procs `json:"procs"` //进程定义
		Check int   `json:"check"` //检查时间间隔（秒）
		log   logger
	}
)

func (g *Guardian) Initialize(log logger) (err error) {
	err = g.Slots.validate()
	if err != nil {
		return err
	}
	err = g.Procs.validate(log)
	if err != nil {
		return err
	}
	if g.Check <= 0 {
		g.Check = 10 //默认10秒检查一次
	}
	g.log = func(level int, msg string, args ...interface{}) {
		msg = fmt.Sprintf("ProcGuard: %s", msg)
		log(level, msg, args...)
	}
	return nil
}

func (g *Guardian) Run() {
	go func() {
		for {
			now := time.Now().Format("15:04:05")
			tags := g.Slots.current(now)
			g.log(1, "now=%s; slots=%+v", now, tags)
			for _, t := range tags {
				for _, p := range g.Procs {
					name := filepath.Base(p.Cmd)
					switch p.Plan[t] {
					case "start":
						if p.done {
							g.log(2, "now=%s; slot=%s; already processed", now, t)
						} else {
							g.log(2, "now=%s; slot=%s; start '%s'", now, t, name)
							if err := p.start(); err == nil {
								p.done = true
							}
						}
					case "stop":
						g.log(2, "now=%s; slot=%s; terminate '%s'", now, t, name)
						p.stop()
					case "keepalive":
						g.log(2, "now=%s; slot=%s; keepalive '%s'", now, t, name)
						p.start()
					default:
						g.log(2, "now=%s; slot=%s; no action for '%s'", now, t, name)
					}
				}
			}
			time.Sleep(time.Duration(g.Check) * time.Second)
		}
	}()
}
