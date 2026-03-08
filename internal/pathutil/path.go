package pathutil

import (
	"path/filepath"
)

// ResolveAbsolute returns a clean absolute path. No workspace restriction;
// paths may be anywhere in the container.
func ResolveAbsolute(path string) (string, error) {
	return filepath.Abs(path)
}
