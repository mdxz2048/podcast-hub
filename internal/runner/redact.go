package runner

import (
	"regexp"
	"strings"
)

type redactionRule struct {
	pattern     *regexp.Regexp
	replacement string
}

var redactionPatterns = []redactionRule{
	{regexp.MustCompile(`(?i)("?(authorization|cookie|password|secret|session|token)"?\s*[:=]\s*")[^"]*"`), `${1}[redacted]"`},
	{regexp.MustCompile(`(?i)(authorization|cookie|password|secret|session|token)(["'\s:=]+)[^,\s}"']+`), `${1}${2}[redacted]`},
	{regexp.MustCompile(`(?i)bearer\s+[a-z0-9._~+/=-]+`), `[redacted]`},
}

func redact(value string) string {
	redacted := value
	for _, rule := range redactionPatterns {
		redacted = rule.pattern.ReplaceAllString(redacted, rule.replacement)
	}
	for _, marker := range []string{"/Users/", ".local/", "work/output", "work/input"} {
		redacted = strings.ReplaceAll(redacted, marker, "[path]/")
	}
	return redacted
}
