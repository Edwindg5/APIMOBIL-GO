package reportgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kajve/api-mobile/internal/domain/entities"
)

// Generator orquesta la recolección de datos de un lote y la construcción
// del archivo PDF o Excel correspondiente, escribiéndolo en disco.
type Generator struct {
	collector  *Collector
	reportsDir string
}

func NewGenerator(collector *Collector, reportsDir string) *Generator {
	return &Generator{collector: collector, reportsDir: reportsDir}
}

// Extension devuelve la extensión de archivo para un formato de reporte.
func Extension(formato string) string {
	if formato == "excel" {
		return "xlsx"
	}
	return "pdf"
}

// FileName construye el nombre determinístico del archivo de un reporte,
// usado tanto al generarlo como al servirlo para descarga.
func FileName(reporteID int, formato string) string {
	return fmt.Sprintf("reporte_%d.%s", reporteID, Extension(formato))
}

// Generate recolecta los datos del lote, arma el PDF o Excel correspondiente
// y lo escribe en disco dentro de reportsDir. Devuelve la ruta del archivo.
func (g *Generator) Generate(ctx context.Context, reporteID int, lote *entities.LoteCafe, usuarioNombre, tipoReporte, formato string) (string, error) {
	data, err := g.collector.Collect(ctx, lote, usuarioNombre, tipoReporte, formato)
	if err != nil {
		return "", fmt.Errorf("error recolectando datos del reporte: %w", err)
	}

	var contenido []byte
	if formato == "excel" {
		contenido, err = BuildExcel(data)
	} else {
		contenido, err = BuildPDF(data)
	}
	if err != nil {
		return "", fmt.Errorf("error generando archivo: %w", err)
	}

	if err := os.MkdirAll(g.reportsDir, 0o755); err != nil {
		return "", fmt.Errorf("error creando directorio de reportes: %w", err)
	}

	path := filepath.Join(g.reportsDir, FileName(reporteID, formato))
	if err := os.WriteFile(path, contenido, 0o644); err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return path, nil
}
