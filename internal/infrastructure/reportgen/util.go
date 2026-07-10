package reportgen

import (
	"fmt"
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

// formatFechaPtr formatea una fecha nullable; fallback es el texto a mostrar
// cuando t es nil (ej. "En curso" para fecha_fin_secado, "Sin iniciar" para
// fecha_inicio_secado — el significado de "nil" depende de la columna).
func formatFechaPtr(t *time.Time, fallback string) string {
	if t == nil {
		return fallback
	}
	return t.Format("02/01/2006 15:04")
}

func capitalizePtr(s *string) string {
	if s == nil {
		return "Sin especificar"
	}
	return capitalize(*s)
}

// textPtr devuelve el valor de un *string sin transformarlo (a diferencia de
// capitalizePtr), para campos de texto libre como la ubicación.
func textPtr(s *string) string {
	if s == nil {
		return "Sin especificar"
	}
	return *s
}

// formatPesoPtr formatea un peso nullable; unit se agrega solo si hay valor
// (ej. " kg"), para no producir "Sin especificar kg".
func formatPesoPtr(p *float64, unit string) string {
	if p == nil {
		return "Sin especificar"
	}
	return fmt.Sprintf("%.1f%s", *p, unit)
}
