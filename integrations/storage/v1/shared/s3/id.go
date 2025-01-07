package s3

import (
	"strings"

	"github.com/dchest/uniuri"
)

// newID creates a new unique ID.
func newID() string {
	return strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
}
