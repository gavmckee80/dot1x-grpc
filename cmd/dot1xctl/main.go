package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	var (
		status  = flag.Bool("status", false, "Show dot1x service status")
		logs    = flag.Bool("logs", false, "Stream journal logs for dot1x service")
		restart = flag.Bool("restart", false, "Restart dot1x service")
		enable  = flag.Bool("enable", false, "Enable and start dot1x service")
		disable = flag.Bool("disable", false, "Disable and stop dot1x service")
	)
	flag.Parse()

	switch {
	case *status:
		run("systemctl", "status", "dot1x.service")
	case *logs:
		run("journalctl", "-u", "dot1x.service", "-f")
	case *restart:
		run("systemctl", "restart", "dot1x.service")
	case *enable:
		run("systemctl", "enable", "--now", "dot1x.service")
	case *disable:
		run("systemctl", "disable", "--now", "dot1x.service")
	default:
		fmt.Println("Usage: dot1xctl [-status|-logs|-restart|-enable|-disable]")
	}
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed: %v", err)
	}
}
