package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

// A label names one run of the service: the data is served only for the
// host <label>.maintainer.localhost, which nobody could serve before the run
// made the label. The label is the credential, so it is in no log line and in
// no error text.

// LabelLen is the length of a label: 16 random bytes as hex.
const LabelLen = 32

// NewLabel returns a fresh label.
func NewLabel() (string, error) {
	raw := make([]byte, LabelLen/2)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("make the access label: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

// ValidLabel reports whether s has the form of a label: 32 lowercase hex
// characters.
func ValidLabel(s string) bool {
	if len(s) != LabelLen {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// LabelHost returns the host name the run with label serves its data on.
func LabelHost(label string) string { return label + "." + Host }

// LabelURL returns the address of the run with label on port.
func LabelURL(label string, port int) string {
	return "http://" + LabelHost(label) + ":" + strconv.Itoa(port) + "/"
}

// Proof answers a nonce for label: HMAC-SHA256 keyed by the label over the
// bytes of the nonce, as hex. It tells the command line the service that
// knows the label from another listener on the port.
func Proof(label string, nonce []byte) string {
	mac := hmac.New(sha256.New, []byte(label))
	mac.Write(nonce)
	return hex.EncodeToString(mac.Sum(nil))
}

// ProofEqual compares two proofs in constant time.
func ProofEqual(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

// StateDir returns the folder of the label files:
// $XDG_STATE_HOME/maintainer/service, else ~/.local/state/maintainer/service.
func StateDir(getenv func(string) string, home string) string {
	base := getenv("XDG_STATE_HOME")
	if base == "" && home != "" {
		base = filepath.Join(home, ".local", "state")
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "maintainer", "service")
}

// LabelPath returns the label file of port in dir.
func LabelPath(dir string, port int) string {
	return filepath.Join(dir, strconv.Itoa(port)+".label")
}

// PrepareStateDir creates dir with mode 0700 and checks that it is a folder,
// not a link, owned by the user and closed to everybody else.
func PrepareStateDir(dir string) error {
	if dir == "" {
		return errors.New("no state folder: set XDG_STATE_HOME or HOME")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create the state folder: %w", err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("check the state folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("the state folder %s is not a folder", dir)
	}
	if !ownedByUser(info) {
		return fmt.Errorf("the state folder %s belongs to another user", dir)
	}
	if info.Mode().Perm() != 0o700 {
		if err := os.Chmod(dir, 0o700); err != nil {
			return fmt.Errorf("restrict the state folder: %w", err)
		}
	}
	return nil
}

// WriteLabel writes label as the label file of port in dir: a temporary file
// created exclusively with mode 0600, renamed over a stale file.
func WriteLabel(dir string, port int, label string) error {
	if err := PrepareStateDir(dir); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".label-*")
	if err != nil {
		return fmt.Errorf("write the label file: %w", err)
	}
	name := tmp.Name()
	_, werr := tmp.WriteString(label + "\n")
	if werr == nil {
		werr = tmp.Chmod(0o600)
	}
	if werr == nil {
		werr = tmp.Sync()
	}
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(name, LabelPath(dir, port))
	}
	if werr != nil {
		_ = os.Remove(name)
		return fmt.Errorf("write the label file: %w", werr)
	}
	return nil
}

// ErrNoLabel means there is no valid label file for the port.
var ErrNoLabel = errors.New("no label file")

// ReadLabel reads the label file of port in dir and validates it. A missing
// file or one without a label is ErrNoLabel.
func ReadLabel(dir string, port int) (string, error) {
	raw, err := os.ReadFile(LabelPath(dir, port))
	if errors.Is(err, fs.ErrNotExist) {
		return "", ErrNoLabel
	}
	if err != nil {
		return "", fmt.Errorf("read the label file: %w", err)
	}
	label := string(bytes.TrimSpace(raw))
	if !ValidLabel(label) {
		return "", ErrNoLabel
	}
	return label, nil
}

// RemoveLabel removes the label file of port in dir if it still holds label:
// a later run that replaced it keeps its file.
func RemoveLabel(dir string, port int, label string) error {
	current, err := ReadLabel(dir, port)
	if err != nil {
		if errors.Is(err, ErrNoLabel) {
			return nil
		}
		return err
	}
	if !ProofEqual(current, label) {
		return nil
	}
	if err := os.Remove(LabelPath(dir, port)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove the label file: %w", err)
	}
	return nil
}
