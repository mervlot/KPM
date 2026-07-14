package resolver

import (
	"errors"
	"testing"
)

func TestNotFoundInAnyRepoErrorMessage(t *testing.T) {
	e := &NotFoundInAnyRepoError{Coordinate: "org.example:foo:1.0", Kind: "dependency", Ext: "pom", RepoCount: 2}
	msg := e.Error()
	if !contains(msg, "not found") || !contains(msg, "org.example:foo:1.0") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestNoConnectionErrorUnwraps(t *testing.T) {
	inner := errors.New("dial tcp: lookup repo1.maven.org: no such host")
	e := &NoConnectionError{Coordinate: "org.example:foo:1.0", Err: inner}
	if !errors.Is(e, inner) {
		t.Error("expected NoConnectionError to unwrap to the inner error")
	}
}

func TestIsConnectivityError(t *testing.T) {
	cases := map[string]bool{
		"dial tcp: lookup repo1.maven.org: no such host": true,
		"connection refused":                             true,
		"repository returned 404 for https://x":          false,
		"repository returned 429 for https://x":          false,
	}
	for msg, want := range cases {
		got := isConnectivityError(errors.New(msg))
		if got != want {
			t.Errorf("isConnectivityError(%q) = %v, want %v", msg, got, want)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}