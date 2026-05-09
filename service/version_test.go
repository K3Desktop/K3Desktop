package service

import (
	"testing"
)

func TestK3sExclude_Regex(t *testing.T) {
	shouldExclude := []string{
		"sha-abc123",
		"v1.28.0-rc1",
		"v1.28.0-alpha1",
		"v1.28.0-beta1",
		"v1.28.0-dev",
		"v1.28.0-test",
		"v1.28.0-arm64",
		"v1.28.0-arm",
		"v1.28.0-amd64",
		"v1.28.0-dind",
		"v1.28.0-engine",
	}

	for _, tag := range shouldExclude {
		if !k3sExclude.MatchString(tag) {
			t.Errorf("k3sExclude should match %q", tag)
		}
	}

	shouldInclude := []string{
		"v1.28.0-k3s1",
		"v1.27.3-k3s1",
		"v1.26.0+k3s1",
		"v1.25.0-k3s2",
	}

	for _, tag := range shouldInclude {
		if k3sExclude.MatchString(tag) {
			t.Errorf("k3sExclude should NOT match %q", tag)
		}
	}
}

func TestK3sExclude_ShaPrefix(t *testing.T) {
	if !k3sExclude.MatchString("sha-0123456789abcdef") {
		t.Error("should exclude sha- prefix tags")
	}
}

func TestK3sExclude_S390x(t *testing.T) {
	if !k3sExclude.MatchString("v1.28.0-k3s1-s390x") {
		t.Error("should exclude s390x tags")
	}
}
