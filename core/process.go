package core

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	addr, host, port string
	flags, env       []string
	exePath, userDir string
	chromeCmd        *exec.Cmd
	apiEndpoint      string
	readyCh          chan struct{}
}

func NewProcess() *Process {
	c := &Process{}
	c.host = "127.0.0.1"
	c.port = "9222"
	c.readyCh = make(chan struct{})
	c.flags = make([]string, 0)
	c.env = make([]string, 0)
	return c
}

func (c *Process) SetExePath(path string) {
	c.exePath = path
}

func (c *Process) SetUserDir(path string) {
	c.userDir = path
}

func (c *Process) AddFlags(flags []string) {
	c.flags = append(c.flags, flags...)
}

func (c *Process) AddEnvironmentVars(vars []string) {
	c.env = append(c.env, vars...)
}

func (c *Process) Start() {
	c.addr = fmt.Sprintf("%s:%s", c.host, c.port)
	c.apiEndpoint = fmt.Sprintf("http://%s/json", c.addr)

	c.flags = append(c.flags, fmt.Sprintf("--user-data-dir=%s", c.userDir))

	c.flags = append(c.flags, fmt.Sprintf("--remote-debugging-port=%s", c.port))

	c.flags = append(c.flags, "--no-first-run")

	c.flags = append(c.flags, "--no-default-browser-check")

	c.chromeCmd = exec.Command(c.exePath, c.flags...)

	c.chromeCmd.Env = os.Environ()
	c.chromeCmd.Env = append(c.chromeCmd.Env, c.env...)
	go func() {
		err := c.chromeCmd.Start()
		if err != nil {
			log.Fatalf("error starting chrome process: %s", err)
		}
		err = c.chromeCmd.Wait()

		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	go c.probeDebugPort()
	<-c.readyCh
}

func (c *Process) Exit() error {
	return c.chromeCmd.Process.Kill()
}

func (c *Process) probeDebugPort() {
	ticker := time.NewTicker(time.Millisecond * 100)
	timeoutTicker := time.NewTicker(time.Second * 15)

	defer func() {
		ticker.Stop()
		timeoutTicker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			resp, err := http.Get(c.apiEndpoint)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			c.readyCh <- struct{}{}
			return
		case <-timeoutTicker.C:
			log.Fatalf("Unable to contact debugger at %s after 15 seconds, gave up", c.apiEndpoint)
		}
	}
}
