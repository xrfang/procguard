package procguard

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type (
	procs []*proc
	proc  struct {
		Cmd  string            `json:"cmd"`  //命令，相对路径为当前程序路径
		Args []string          `json:"args"` //参数
		Plan map[string]string //执行计划，key为slots的key，value为所需状态：start, stop或keepalive
		pid  int               //子进程ID
		done bool              //是否已经执行启动动作
		cmdl string
		log  logger
	}
)

func (p *proc) validate(log logger) error {
	for s, x := range p.Plan {
		x = strings.ToLower(x)
		switch x {
		case "start", "stop", "keepalive":
			p.Plan[s] = x
		default:
			return fmt.Errorf("invalid action '%s'", x)
		}
	}
	if p.Cmd == "" {
		return errors.New("proc: empty command line")
	}
	if !filepath.IsAbs(p.Cmd) {
		xp, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		p.Cmd = filepath.Join(xp, p.Cmd)
	}
	name := filepath.Base(p.Cmd)
	p.cmdl = strings.Join(append([]string{p.Cmd}, p.Args...), string(0))
	hash := crc32.ChecksumIEEE([]byte(p.cmdl))
	p.log = func(level int, msg string, args ...interface{}) {
		msg = fmt.Sprintf("[%08x]ProcGuard(%s): %s", hash, name, msg)
		log(level, msg, args...)
	}
	return nil
}

func (p *proc) killByCmdLine() int {
	return 0 //TODO...
}

func (p *proc) start() (err error) {
	var process *os.Process
	if p.pid != 0 {
		process, err = os.FindProcess(p.pid)
		if err == nil {
			err = process.Signal(syscall.Signal(0))
		}
		if err == nil {
			p.log(2, "pid %d already running", p.pid)
			return
		}
		p.log(0, "pid %d gone, will restart", p.pid)
		p.pid = 0
	}
	if cnt := p.killByCmdLine(); cnt > 0 {
		p.log(0, "killed dangling instance before start")
	}
	var buf bytes.Buffer
	cmd := exec.Command(p.Cmd, p.Args...)
	cmd.Dir = filepath.Dir(p.Cmd)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Start()
	go func() {
		err := cmd.Wait()
		p.log(0, "terminated, err=%v", err)
		if err != nil {
			output := strings.TrimSpace(buf.String())
			if len(output) > 0 {
				p.log(1, output)
			}
		}
	}()
	if err == nil {
		p.pid = cmd.Process.Pid
		p.log(0, "started, pid=%d", p.pid)
	} else {
		p.log(0, "start failed, err=%v", err)
	}
	return
}

func (p *proc) stop() (err error) {
	var process *os.Process
	defer func() {
		if p.pid == 0 {
			return
		}
		if err == nil {
			p.log(0, "killed pid %d", p.pid)
		} else {
			p.log(0, "kill pid %d failed, err=%v", p.pid, err)
		}
		p.pid = 0
	}()
	process, err = os.FindProcess(p.pid)
	if err == nil {
		err = process.Kill()
	}
	if err == nil {
		return
	}
	if cnt := p.killByCmdLine(); cnt > 0 {
		p.pid = 0
		p.log(0, "killed %d instances by name", cnt)
	} else {
		p.log(2, "no process killed")
	}
	return
}

func (ps procs) validate(log logger) error {
	for _, p := range ps {
		if err := p.validate(log); err != nil {
			return err
		}
	}
	return nil
}
