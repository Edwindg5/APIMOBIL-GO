package reportgen

import (
	"strings"
	"time"
)

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func formatFechaPtr(t *time.Time) string {
	if t == nil {
		return "En curso"
	}
	return t.Format("02/01/2006 15:04")
}

func capitalizePtr(s *string) string {
	if s == nil {
		return "Sin especificar"
	}
	return capitalize(*s)
}
