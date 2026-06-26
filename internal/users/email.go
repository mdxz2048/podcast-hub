package users

import (
	"errors"
	"strings"
)

func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", errors.New("email is required")
	}
	if !strings.Contains(normalized, "@") || strings.HasPrefix(normalized, "@") || strings.HasSuffix(normalized, "@") {
		return "", errors.New("invalid email")
	}
	return normalized, nil
}

func EmailHint(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	local := parts[0]
	if len(local) == 1 {
		return "*" + "@" + parts[1]
	}
	return local[:1] + "***@" + parts[1]
}
