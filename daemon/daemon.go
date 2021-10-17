package daemon

import (
	"fmt"
	"os/exec"
	"time"
)

type Daemon struct {
	Registry      string   // Docker registry
	Mirror        string   // Docker registry mirror
	Insecure      bool     // Docker daemon enable insecure registries
	StorageDriver string   // Docker daemon storage driver
	StoragePath   string   // Docker daemon storage path
	Disabled      bool     // Docker daemon is disabled (already running)
	Debug         bool     // Docker daemon started in debug mode
	Bip           string   // Docker daemon network bridge IP address
	DNS           []string // Docker daemon dns server
	DNSSearch     []string // Docker daemon dns search domain
	MTU           string   // Docker daemon mtu setting
	IPv6          bool     // Docker daemon IPv6 networking
	Experimental  bool     // Docker daemon enable experimental mode
}

func StartDaemon(d Daemon) error {
	if !d.Disabled {
		startDaemon(d)
	}
	return waitForDaemon()
}

func waitForDaemon() error {
	// poll the docker daemon until it is started. This ensures the daemon is
	// ready to accept connections before we proceed.
	for i := 0; ; i++ {
		cmd := commandInfo()
		err := cmd.Run()
		if err == nil {
			break
		}
		if i == 15 {
			fmt.Println("Unable to reach Docker Daemon after 15 attempts.")
			return fmt.Errorf("failed to reach docker daemon after 15 attempts: %v", err)
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

// helper function to create the docker info command.
func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}
