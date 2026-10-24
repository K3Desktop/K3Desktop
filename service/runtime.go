package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
		exportDockerSock()
		dockerRT = dockerrt.Docker{}
	})
	return dockerRT
}

// resolveDockerHost probes available container engine endpoints and sets DOCKER_HOST.
// Covers default Docker, rootless/rootful Podman sockets, and Docker-in-WSL2 with
// an exposed TCP socket on Windows.
func resolveDockerHost() {
	if probeDockerHost("") {
		return
	}
	for _, host := range podmanHosts() {
		if probeDockerHost(host) {
			os.Setenv("DOCKER_HOST", host)
			return
		}
	}
	if probeDockerHost("tcp://localhost:2375") {
		os.Setenv("DOCKER_HOST", "tcp://localhost:2375")
	}
}

func podmanHosts() []string {
	var hosts []string
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		hosts = append(hosts, "unix://"+filepath.Join(dir, "podman", "podman.sock"))
	}
	hosts = append(hosts,
		fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", os.Getuid()),
		"unix:///run/podman/podman.sock",
	)
	return hosts
}

// exportDockerSock points k3d's DOCKER_SOCK at the active unix socket so the
// k3d-tools container bind-mounts a path that exists (default is /var/run/docker.sock).
func exportDockerSock() {
	if os.Getenv("DOCKER_SOCK") != "" {
		return
	}
	if sock, ok := strings.CutPrefix(os.Getenv("DOCKER_HOST"), "unix://"); ok && sock != "" {
		os.Setenv("DOCKER_SOCK", sock)
	}
}

// isRootlessEngine reports whether the container engine runs in a user namespace.
func isRootlessEngine() bool {
	GetRuntime()
	c, err := dockerrt.GetDockerClient()
	if err != nil {
		return false
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	info, err := c.Info(ctx)
	if err != nil {
		return false
	}
	return slices.Contains(info.SecurityOptions, "name=rootless")
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
