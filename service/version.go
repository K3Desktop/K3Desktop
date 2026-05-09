package service

import (
	"context"
	"regexp"
	"sort"
	"github.com/google/go-containerregistry/pkg/crane"
)

type VersionService struct{}

var (
	k3sExclude = regexp.MustCompile(`^sha-|.+(rc|engine|dind|alpha|beta|dev|test|arm|arm64|amd64|s390x).*`)
)

func (s *VersionService) ListK3sVersions(ctx context.Context, limit int) ([]string, error) {
	tags, err := crane.ListTags("docker.io/rancher/k3s")
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(tags))
	for _, t := range tags {
		if !k3sExclude.MatchString(t) {
			filtered = append(filtered, t)
		}
	}

	// descending sort: newer versions first
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i] > filtered[j]
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}
