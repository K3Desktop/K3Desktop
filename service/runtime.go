package service

import (
	"sync"

	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	dockerrt "github.com/k3d-io/k3d/v5/pkg/runtimes/docker"
)

var (
	once     sync.Once
	dockerRT runtimes.Runtime
)

func GetRuntime() runtimes.Runtime {
	once.Do(func() { dockerRT = dockerrt.Docker{} })
	return dockerRT
}
