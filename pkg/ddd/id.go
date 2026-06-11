package ddd

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ID string

// ParseID validates an existing domain identity.
func ParseID(raw string) (ID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("ddd id is required")
	}
	return ID(raw), nil
}

// NewID is kept for compatibility. Prefer ParseID when validating an existing
// identity and IDGenerator.NewID when creating a new identity.
func NewID(raw string) (ID, error) {
	return ParseID(raw)
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsZero() bool {
	return strings.TrimSpace(string(id)) == ""
}

type IDGenerator interface {
	NewID(prefix string) (ID, error)
}

type UUIDGenerator struct{}

var _ IDGenerator = (*UUIDGenerator)(nil)

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) NewID(prefix string) (ID, error) {
	prefix = normalizeIDPrefix(prefix)
	value := uuid.NewString()
	if prefix != "" {
		value = prefix + "_" + value
	}
	return ID(value), nil
}

func normalizeIDPrefix(prefix string) string {
	prefix = strings.TrimSpace(strings.ToLower(prefix))
	if prefix == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range prefix {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	return out
}
