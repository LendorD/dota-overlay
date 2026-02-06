package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

type childProc struct {
	name string
	cmd  *exec.Cmd
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	overlayCmd, err := buildCommand("overlay", root, []string{"./cmd/overlay"})
	if err != nil {
		log.Fatal(err)
	}
	dotaplusCmd, err := buildCommand("dotaplus", root, []string{"./cmd/dotaplus"})
	if err != nil {
		log.Fatal(err)
	}

	procs := []childProc{
		{name: "overlay", cmd: overlayCmd},
		{name: "dotaplus", cmd: dotaplusCmd},
	}

	for _, p := range procs {
		p.cmd.Stdout = os.Stdout
		p.cmd.Stderr = os.Stderr
		p.cmd.Stdin = os.Stdin
		if err := p.cmd.Start(); err != nil {
			log.Fatalf("failed to start %s: %v", p.name, err)
		}
		log.Printf("started %s (pid %d)", p.name, p.cmd.Process.Pid)
	}

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(len(procs))

	for _, p := range procs {
		proc := p
		go func() {
			defer wg.Done()
			if err := proc.cmd.Wait(); err != nil {
				log.Printf("%s exited: %v", proc.name, err)
			} else {
				log.Printf("%s exited", proc.name)
			}
		}()
	}

	select {
	case <-stop:
		log.Printf("shutdown requested")
		for _, p := range procs {
			_ = terminateProcess(p.cmd)
		}
	}

	wg.Wait()
}

func buildCommand(name string, workdir string, goRunArgs []string) (*exec.Cmd, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)

	binName := name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	candidate := filepath.Join(exeDir, binName)
	if _, err := os.Stat(candidate); err == nil {
		cmd := exec.Command(candidate)
		cmd.Dir = exeDir
		return cmd, nil
	}

	cmd := exec.Command("go", append([]string{"run"}, goRunArgs...)...)
	cmd.Dir = workdir
	return cmd, nil
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		return cmd.Process.Kill()
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return cmd.Process.Kill()
	}

	return nil
}
