// Package checksum computes and verifies artifact checksums against the
// sidecar hash files Maven repositories publish alongside every artifact
// (foo.jar.sha1, foo.jar.md5, and increasingly foo.jar.sha256).
package checksum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type Algo string

const (
	SHA256 Algo = "sha256"
	SHA1   Algo = "sha1"
	MD5    Algo = "md5"
)

// Extension returns the sidecar file suffix Maven repos use for this algo.
func (a Algo) Extension() string {
	switch a {
	case SHA256:
		return ".sha256"
	case SHA1:
		return ".sha1"
	case MD5:
		return ".md5"
	}
	return ""
}

// PreferredOrder is the order in which we attempt checksum verification.
// SHA1 is tried first deliberately: real-world Maven repositories (Central
// included) publish .sha1/.md5 sidecars far more consistently than .sha256,
// so trying SHA256 first means eating a near-guaranteed 404 (and the
// network round-trip + rate-limit "cost" of that request) on the vast
// majority of artifacts before falling back to the one that actually
// exists. This is a request-volume/security trade-off: SHA1 is still fine
// for tamper/corruption *detection* against a hash the repository itself
// published (the threat model here isn't an attacker crafting a SHA1
// collision, it's "did this download get truncated/corrupted/MITM'd
// against what the repo says it should be"). Callers that need
// cryptographically strong verification for a specific high-value artifact
// can still request SHA256 explicitly.
var PreferredOrder = []Algo{SHA1, SHA256, MD5}

// HashFile computes the checksum of a file on disk using the given algorithm.
func HashFile(path string, algo Algo) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return HashReader(f, algo)
}

// HashReader computes the checksum of a stream using the given algorithm.
func HashReader(r io.Reader, algo Algo) (string, error) {
	var h interface {
		io.Writer
		Sum([]byte) []byte
	}
	switch algo {
	case SHA256:
		h = sha256.New()
	case SHA1:
		h = sha1.New()
	case MD5:
		h = md5.New()
	default:
		return "", fmt.Errorf("unsupported checksum algo: %s", algo)
	}
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ParseSidecar extracts the hex digest from a downloaded sidecar file's
// content, tolerating both bare-hex files and "hexdigest  filename" formats.
func ParseSidecar(content []byte) string {
	s := strings.TrimSpace(string(content))
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	// Some servers emit "<hex> <filename>"; the hex is always the first
	// token and is the only thing composed entirely of hex digits.
	for _, f := range fields {
		if isHex(f) {
			return strings.ToLower(f)
		}
	}
	return strings.ToLower(fields[0])
}

func isHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// Verify computes the checksum of the file at path and compares it against
// expectedHex (case-insensitive). Returns a descriptive error on mismatch.
func Verify(path string, algo Algo, expectedHex string) error {
	got, err := HashFile(path, algo)
	if err != nil {
		return fmt.Errorf("checksum: could not hash %s: %w", path, err)
	}
	if !strings.EqualFold(got, expectedHex) {
		return &MismatchError{Path: path, Algo: algo, Expected: expectedHex, Actual: got}
	}
	return nil
}

// MismatchError is returned when a downloaded artifact's checksum does not
// match the value published by the repository — a strong signal of a
// corrupted download, a MITM'd mirror, or a compromised/tampered artifact.
type MismatchError struct {
	Path     string
	Algo     Algo
	Expected string
	Actual   string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf(
		"checksum mismatch for %s (%s): expected %s, got %s — the download may be corrupt or the repository/mirror may be untrustworthy; re-run with --refresh or verify the repository URL",
		e.Path, e.Algo, e.Expected, e.Actual,
	)
}