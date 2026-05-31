package toolkit

import "strings"

const defaultIdentifier = "tenant"

// Identifier returns a deterministic lowercase identifier safe for Postgres schema/db names.
func Identifier(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
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
	if out == "" {
		return defaultIdentifier
	}
	if out[0] >= '0' && out[0] <= '9' {
		return defaultIdentifier + "_" + out
	}
	return out
}

func SchemaName(prefix string, tenantID string) string {
	if prefix == "" {
		prefix = "t_"
	}
	return prefix + Identifier(tenantID)
}
