package service

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	dockerrt "github.com/k3d-io/k3d/v5/pkg/runtimes/docker"
)

var (
	once     sync.Once
	dockerRT runtimes.Runtime
)

func GetRuntime() runtimes.Runtime {
	once.Do(func() {
		if os.Getenv("DOCKER_HOST") == "" {
			resolveDockerHost()
		}
		dockerRT = dockerrt.Docker{}
	})
	return dockerRT
}

// resolveDockerHost probes available Docker endpoints and sets DOCKER_HOST.
// TCP localhost fallback covers Docker-in-WSL2 with exposed socket on Windows.
func resolveDockerHost() {
	if probeDockerHost("") {
		return
	}
	if probeDockerHost("tcp://localhost:2375") {
		os.Setenv("DOCKER_HOST", "tcp://localhost:2375")
	}
}

func probeDockerHost(host string) bool {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	c, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return false
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = c.Ping(ctx)
	return err == nil
}
