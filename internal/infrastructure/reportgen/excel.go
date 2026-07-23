package reportgen

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

var colLetters = []string{"A", "B", "C", "D", "E", "F", "G"}

func cellRef(col, row int) string {
	return fmt.Sprintf("%s%d", colLetters[col], row)
}

// BuildExcel genera el archivo .xlsx de un reporte de lote, con la misma
// identidad visual (encabezados oscuros + dorado) que el resto de la app.
func BuildExcel(data *ReportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	e, err := newExcelBuilder(f, data)
	if err != nil {
		return nil, err
	}

	if err := e.buildResumen(); err != nil {
		return nil, err
	}
	if err := e.buildLecturas(); err != nil {
		return nil, err
	}
	if err := e.buildAlertas(); err != nil {
		return nil, err
	}
	if err := e.buildPrediccionesYRecomendaciones(); err != nil {
		return nil, err
	}
	if err := e.buildHistorial(); err != nil {
		return nil, err
	}

	if err := f.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}
	if idx, err := f.GetSheetIndex("Resumen"); err == nil {
		f.SetActiveSheet(idx)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type excelBuilder struct {
	f    *excelize.File
	data *ReportData

	styleTitle    int
	styleSubtitle int
	styleHeader   int
	styleLabel    int
	styleValue    int
	styleBody     int
	styleBodyAlt  int
	styleBodyNum  int
	styleBodyNumA int
	styleWrap     int
	styleWrapAlt  int
	styleDanger   int
	styleWarning  int
}

func newExcelBuilder(f *excelize.File, data *ReportData) (*excelBuilder, error) {
	e := &excelBuilder{f: f, data: data}

	styles := []struct {
		id *int
		st *excelize.Style
	}{
		{&e.styleTitle, &excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 18, Color: hexGold},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexBgDark}, Pattern: 1},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
		}},
		{&e.styleSubtitle, &excelize.Style{
			Font:      &excelize.Font{Italic: true, Size: 10, Color: hexGoldSoft},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexBgDark}, Pattern: 1},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
		}},
		{&e.styleHeader, &excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 10, Color: hexTextDark},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexGold}, Pattern: 1},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "center", WrapText: true},
			Border: []excelize.Border{
				{Type: "bottom", Color: hexTextDark, Style: 1},
			},
		}},
		{&e.styleLabel, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: "8C7A69"},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexRowAlt}, Pattern: 1},
		}},
		{&e.styleValue, &excelize.Style{
			Font:      &excelize.Font{Bold: true, Size: 10.5, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
		}},
		{&e.styleBody, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexWhite}, Pattern: 1},
		}},
		{&e.styleBodyAlt, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexRowAlt}, Pattern: 1},
		}},
		{&e.styleBodyNum, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "right", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexWhite}, Pattern: 1},
			NumFmt:    2,
		}},
		{&e.styleBodyNumA, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "right", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexRowAlt}, Pattern: 1},
			NumFmt:    2,
		}},
		{&e.styleWrap, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "top", Horizontal: "left", WrapText: true, Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexWhite}, Pattern: 1},
		}},
		{&e.styleWrapAlt, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Color: hexTextDark},
			Alignment: &excelize.Alignment{Vertical: "top", Horizontal: "left", WrapText: true, Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{hexRowAlt}, Pattern: 1},
		}},
		{&e.styleDanger, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Bold: true, Color: hexDanger},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"FBE4E1"}, Pattern: 1},
		}},
		{&e.styleWarning, &excelize.Style{
			Font:      &excelize.Font{Size: 10, Bold: true, Color: hexWarning},
			Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "left", Indent: 1},
			Fill:      excelize.Fill{Type: "pattern", Color: []string{"FBF0DA"}, Pattern: 1},
		}},
	}

	for _, s := range styles {
		id, err := f.NewStyle(s.st)
		if err != nil {
			return nil, err
		}
		*s.id = id
	}

	return e, nil
}

// writeBanner escribe el título de marca en las dos primeras filas de una hoja.
func (e *excelBuilder) writeBanner(sheet, title, subtitle string, lastCol int) error {
	f := e.f
	if err := f.SetRowHeight(sheet, 1, 26); err != nil {
		return err
	}
	if err := f.SetRowHeight(sheet, 2, 18); err != nil {
		return err
	}
	if err := f.MergeCell(sheet, cellRef(0, 1), cellRef(lastCol, 1)); err != nil {
		return err
	}
	if err := f.MergeCell(sheet, cellRef(0, 2), cellRef(lastCol, 2)); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, cellRef(0, 1), title); err != nil {
		return err
	}
	if err := f.SetCellValue(sheet, cellRef(0, 2), subtitle); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, cellRef(0, 1), cellRef(lastCol, 1), e.styleTitle); err != nil {
		return err
	}
	return f.SetCellStyle(sheet, cellRef(0, 2), cellRef(lastCol, 2), e.styleSubtitle)
}

func (e *excelBuilder) writeHeaderRow(sheet string, row int, headers []string) error {
	f := e.f
	if err := f.SetRowHeight(sheet, row, 18); err != nil {
		return err
	}
	for i, h := range headers {
		if err := f.SetCellValue(sheet, cellRef(i, row), h); err != nil {
			return err
		}
	}
	return f.SetCellStyle(sheet, cellRef(0, row), cellRef(len(headers)-1, row), e.styleHeader)
}

func (e *excelBuilder) buildResumen() error {
	f := e.f
	sheet := "Resumen"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "A", 24); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 34); err != nil {
		return err
	}

	subtitle := fmt.Sprintf("Reporte de %s · %s", capitalize(e.data.TipoReporte), e.data.GeneradoEn.Format("02/01/2006 15:04"))
	if err := e.writeBanner(sheet, "KAJVE — Reporte de secado de café", subtitle, 1); err != nil {
		return err
	}

	lote := e.data.Lote
	loteRows := [][2]string{
		{"Lote", lote.NombreLote},
		{"Variedad", capitalizePtr(lote.Variedad)},
		{"Proceso", capitalizePtr(lote.TipoProceso)},
		{"Peso (kg)", formatPesoPtr(lote.PesoKg, "")},
		{"Ubicación", textPtr(lote.Ubicacion)},
		{"Estado", capitalize(lote.Estado)},
		{"Usuario", e.data.UsuarioNombre},
		{"Inicio de secado", formatFechaPtr(lote.FechaInicioSecado, "Sin iniciar")},
		{"Fin de secado", formatFechaPtr(lote.FechaFinSecado, "En curso")},
		{"Código QR", lote.CodigoQR},
	}
	row, err := e.writeKeyValueTable(sheet, 4, "Información del lote", loteRows)
	if err != nil {
		return err
	}

	if st := e.data.Estadisticas; st != nil {
		statRows := [][2]string{
			{"Temperatura promedio (°C)", fmt.Sprintf("%.1f", st.TemperaturaPromedio)},
			{"Temperatura mínima (°C)", fmt.Sprintf("%.1f", st.TemperaturaMin)},
			{"Temperatura máxima (°C)", fmt.Sprintf("%.1f", st.TemperaturaMax)},
			{"Humedad promedio (%)", fmt.Sprintf("%.1f", st.HumedadPromedio)},
			{"Humedad mínima (%)", fmt.Sprintf("%.1f", st.HumedadMin)},
			{"Humedad máxima (%)", fmt.Sprintf("%.1f", st.HumedadMax)},
			{"Días de secado", fmt.Sprintf("%d", st.DiasSecado)},
			{"Total de lecturas", fmt.Sprintf("%d", st.TotalLecturas)},
			{"Total de alertas", fmt.Sprintf("%d", st.TotalAlertas)},
			{"Alertas críticas", fmt.Sprintf("%d", st.AlertasCriticas)},
			{"Alertas sin atender", fmt.Sprintf("%d", st.AlertasSinAtender)},
		}
		if _, err := e.writeKeyValueTable(sheet, row+1, "Indicadores del proceso", statRows); err != nil {
			return err
		}
	}

	return nil
}

func (e *excelBuilder) writeKeyValueTable(sheet string, startRow int, title string, rows [][2]string) (int, error) {
	f := e.f
	if err := f.MergeCell(sheet, cellRef(0, startRow), cellRef(1, startRow)); err != nil {
		return 0, err
	}
	if err := f.SetCellValue(sheet, cellRef(0, startRow), title); err != nil {
		return 0, err
	}
	if err := f.SetCellStyle(sheet, cellRef(0, startRow), cellRef(1, startRow), e.styleHeader); err != nil {
		return 0, err
	}

	row := startRow + 1
	for _, r := range rows {
		if err := f.SetCellValue(sheet, cellRef(0, row), r[0]); err != nil {
			return 0, err
		}
		if err := f.SetCellValue(sheet, cellRef(1, row), r[1]); err != nil {
			return 0, err
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(0, row), e.styleLabel); err != nil {
			return 0, err
		}
		if err := f.SetCellStyle(sheet, cellRef(1, row), cellRef(1, row), e.styleValue); err != nil {
			return 0, err
		}
		row++
	}
	return row, nil
}

func (e *excelBuilder) buildLecturas() error {
	f := e.f
	sheet := "Lecturas"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "A", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "C", 18); err != nil {
		return err
	}

	if err := e.writeBanner(sheet, "Lecturas ambientales", fmt.Sprintf("Lote: %s", e.data.Lote.NombreLote), 2); err != nil {
		return err
	}
	if err := e.writeHeaderRow(sheet, 4, []string{"Fecha y hora", "Temperatura (°C)", "Humedad (%)"}); err != nil {
		return err
	}

	row := 5
	for i, l := range e.data.Lecturas {
		bodyStyle, numStyle := e.styleBody, e.styleBodyNum
		if i%2 == 1 {
			bodyStyle, numStyle = e.styleBodyAlt, e.styleBodyNumA
		}
		if err := f.SetCellValue(sheet, cellRef(0, row), l.Timestamp.Format("02/01/2006 15:04")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(1, row), l.Temperatura); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(2, row), l.Humedad); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(0, row), bodyStyle); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cellRef(1, row), cellRef(2, row), numStyle); err != nil {
			return err
		}
		row++
	}

	return f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, YSplit: 4, TopLeftCell: "A5", ActivePane: "bottomLeft",
	})
}

func (e *excelBuilder) buildAlertas() error {
	f := e.f
	sheet := "Alertas"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "A", 18); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 14); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "C", "C", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "D", "D", 50); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "E", "E", 12); err != nil {
		return err
	}

	if err := e.writeBanner(sheet, "Alertas", fmt.Sprintf("Lote: %s", e.data.Lote.NombreLote), 4); err != nil {
		return err
	}
	if err := e.writeHeaderRow(sheet, 4, []string{"Fecha", "Severidad", "Tipo", "Mensaje", "Atendida"}); err != nil {
		return err
	}

	row := 5
	for _, a := range e.data.Alertas {
		style := e.styleBody
		switch a.NivelSeveridad {
		case "critica":
			style = e.styleDanger
		case "alta", "media":
			style = e.styleWarning
		}
		atendida := "No"
		if a.Atendida {
			atendida = "Sí"
		}
		values := []interface{}{
			a.FechaGenerada.Format("02/01/2006 15:04"),
			strings.ToUpper(a.NivelSeveridad),
			capitalize(strings.ReplaceAll(a.TipoAlerta, "_", " ")),
			a.Mensaje,
			atendida,
		}
		for i, v := range values {
			if err := f.SetCellValue(sheet, cellRef(i, row), v); err != nil {
				return err
			}
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(4, row), style); err != nil {
			return err
		}
		row++
	}

	return f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, YSplit: 4, TopLeftCell: "A5", ActivePane: "bottomLeft",
	})
}

func (e *excelBuilder) buildPrediccionesYRecomendaciones() error {
	if len(e.data.Predicciones) == 0 && len(e.data.Recomendaciones) == 0 {
		return nil
	}
	f := e.f
	sheet := "Predicciones"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "A", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "D", 22); err != nil {
		return err
	}

	if err := e.writeBanner(sheet, "Predicciones y recomendaciones", fmt.Sprintf("Lote: %s", e.data.Lote.NombreLote), 3); err != nil {
		return err
	}

	row := 4
	if len(e.data.Predicciones) > 0 {
		if err := e.writeHeaderRow(sheet, row, []string{"Fecha", "Tiempo estimado (h)", "Calidad estimada", "Confianza (%)"}); err != nil {
			return err
		}
		row++
		for i, p := range e.data.Predicciones {
			style, numStyle := e.styleBody, e.styleBodyNum
			if i%2 == 1 {
				style, numStyle = e.styleBodyAlt, e.styleBodyNumA
			}
			var tiempoVal interface{} = "Pendiente"
			if p.TiempoEstimadoHoras != nil {
				tiempoVal = *p.TiempoEstimadoHoras
			}
			// CalidadEstimada es un puntaje escala SCA 0-100 (ya no una categoría) -- ver
			// microservicioMLL/migration.sql paso 10.
			var calidadVal interface{} = "Pendiente"
			if p.CalidadEstimada != nil {
				calidadVal = *p.CalidadEstimada
			}
			var confVal interface{} = "Pendiente"
			if p.Confianza != nil {
				c := *p.Confianza
				if c <= 1.0 {
					c *= 100
				}
				confVal = c
			}
			if err := f.SetCellValue(sheet, cellRef(0, row), p.FechaPrediccion.Format("02/01/2006 15:04")); err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cellRef(1, row), tiempoVal); err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cellRef(2, row), calidadVal); err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cellRef(3, row), confVal); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(0, row), style); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cellRef(1, row), cellRef(1, row), numStyle); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cellRef(2, row), cellRef(2, row), numStyle); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cellRef(3, row), cellRef(3, row), numStyle); err != nil {
				return err
			}
			row++
		}
		row++
	}

	if len(e.data.Recomendaciones) > 0 {
		if err := f.MergeCell(sheet, cellRef(0, row), cellRef(3, row)); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(0, row), "Recomendaciones"); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(3, row), e.styleHeader); err != nil {
			return err
		}
		row++
		for i, r := range e.data.Recomendaciones {
			style := e.styleWrap
			if i%2 == 1 {
				style = e.styleWrapAlt
			}
			if err := f.SetRowHeight(sheet, row, 30); err != nil {
				return err
			}
			if err := f.MergeCell(sheet, cellRef(0, row), cellRef(3, row)); err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cellRef(0, row), "• "+r.Texto); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(3, row), style); err != nil {
				return err
			}
			row++
		}
	}

	return nil
}

func (e *excelBuilder) buildHistorial() error {
	f := e.f
	sheet := "Historial"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "A", "A", 20); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "B", "B", 24); err != nil {
		return err
	}
	if err := f.SetColWidth(sheet, "C", "C", 60); err != nil {
		return err
	}

	if err := e.writeBanner(sheet, "Historial de eventos", fmt.Sprintf("Lote: %s", e.data.Lote.NombreLote), 2); err != nil {
		return err
	}
	if err := e.writeHeaderRow(sheet, 4, []string{"Fecha", "Evento", "Descripción"}); err != nil {
		return err
	}

	row := 5
	for i, h := range e.data.Historial {
		style := e.styleBody
		wrapStyle := e.styleWrap
		if i%2 == 1 {
			style = e.styleBodyAlt
			wrapStyle = e.styleWrapAlt
		}
		if err := f.SetCellValue(sheet, cellRef(0, row), h.FechaEvento.Format("02/01/2006 15:04")); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(1, row), capitalize(strings.ReplaceAll(h.TipoEvento, "_", " "))); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cellRef(2, row), h.Descripcion); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cellRef(0, row), cellRef(1, row), style); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cellRef(2, row), cellRef(2, row), wrapStyle); err != nil {
			return err
		}
		row++
	}

	return f.SetPanes(sheet, &excelize.Panes{
		Freeze: true, YSplit: 4, TopLeftCell: "A5", ActivePane: "bottomLeft",
	})
}
