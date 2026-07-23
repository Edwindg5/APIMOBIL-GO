package reportgen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

const (
	pageW        = 210.0 // A4 en mm
	marginX      = 15.0
	headerHeight = 26.0
	footerHeight = 16.0
)

// BuildPDF genera el PDF de un reporte de lote con la identidad visual
// de la app (fondo café oscuro + dorado en encabezado/acentos, cuerpo
// claro para que sea legible e imprimible).
func BuildPDF(data *ReportData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, footerHeight)
	pdf.SetMargins(marginX, headerHeight+6, marginX)
	pdf.AliasNbPages("{nb}")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	b := &pdfBuilder{pdf: pdf, tr: tr, data: data}
	pdf.SetHeaderFunc(b.drawHeader)
	pdf.SetFooterFunc(b.drawFooter)

	pdf.AddPage()
	b.drawTitleBlock()
	b.drawLoteInfoCard()
	if data.Estadisticas != nil {
		b.drawKPIRow()
	}
	b.drawLecturasSection()
	b.drawAlertasSection()
	b.drawPrediccionesSection()
	b.drawRecomendacionesSection()
	b.drawHistorialSection()

	if err := pdf.Error(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type pdfBuilder struct {
	pdf  *fpdf.Fpdf
	tr   func(string) string
	data *ReportData
}

func pageBreakTriggerY(pdf *fpdf.Fpdf) float64 {
	pageH, _ := pdf.GetPageSize()
	_, _, _, bMargin := pdf.GetMargins()
	return pageH - bMargin
}

func (b *pdfBuilder) ensureSpace(h float64) {
	if b.pdf.GetY()+h > pageBreakTriggerY(b.pdf) {
		b.pdf.AddPage()
	}
}

func (b *pdfBuilder) drawHeader() {
	pdf := b.pdf
	pdf.SetFillColor(colorBgDark.R, colorBgDark.G, colorBgDark.B)
	pdf.Rect(0, 0, pageW, headerHeight, "F")
	pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.Rect(0, headerHeight-1.2, pageW, 1.2, "F")

	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.SetXY(marginX, 6)
	pdf.CellFormat(100, 9, b.tr("KAJVE"), "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
	pdf.SetXY(marginX, 15)
	pdf.CellFormat(140, 5, b.tr("Sistema de monitoreo de secado de café"), "", 0, "L", false, 0, "")

	badge := strings.ToUpper(b.data.TipoReporte)
	pdf.SetFont("Arial", "B", 10)
	bw := pdf.GetStringWidth(b.tr(badge)) + 14
	if bw < 34 {
		bw = 34
	}
	bx := pageW - marginX - bw
	pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.RoundedRect(bx, 7, bw, 9, 2, "1234", "F")
	pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
	pdf.SetXY(bx, 7)
	pdf.CellFormat(bw, 9, b.tr(badge), "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
	pdf.SetXY(bx, 17)
	pdf.CellFormat(bw, 5, b.tr("Formato "+strings.ToUpper(b.data.Formato)), "", 0, "C", false, 0, "")

	pdf.SetY(headerHeight + 4)
}

func (b *pdfBuilder) drawFooter() {
	pdf := b.pdf
	pdf.SetY(-footerHeight)
	pdf.SetDrawColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.SetLineWidth(0.4)
	pdf.Line(marginX, pdf.GetY(), pageW-marginX, pdf.GetY())

	pdf.SetY(-footerHeight + 3)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(colorMuted.R, colorMuted.G, colorMuted.B)

	pdf.SetX(marginX)
	pdf.CellFormat(80, 5, b.tr("Lote: "+b.data.Lote.NombreLote), "", 0, "L", false, 0, "")

	pdf.SetX(marginX)
	pdf.CellFormat(pageW-2*marginX, 5, b.tr("Generado el "+b.data.GeneradoEn.Format("02/01/2006 15:04")), "", 0, "C", false, 0, "")

	pdf.SetX(pageW - marginX - 40)
	pdf.CellFormat(40, 5, b.tr(fmt.Sprintf("Página %d de {nb}", pdf.PageNo())), "", 0, "R", false, 0, "")
}

func (b *pdfBuilder) drawTitleBlock() {
	pdf := b.pdf
	lote := b.data.Lote

	estadoTxt := strings.ToUpper(lote.Estado)
	pdf.SetFont("Arial", "B", 9)
	ew := pdf.GetStringWidth(b.tr(estadoTxt)) + 10
	ex := pageW - marginX - ew
	y := pdf.GetY()
	pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.RoundedRect(ex, y, ew, 7.5, 1.5, "1234", "F")
	pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
	pdf.SetXY(ex, y)
	pdf.CellFormat(ew, 7.5, b.tr(estadoTxt), "", 0, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
	pdf.SetXY(marginX, y)
	pdf.CellFormat(ex-marginX, 8, b.tr(lote.NombreLote), "", 0, "L", false, 0, "")

	pdf.SetY(y + 10)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(colorMuted.R, colorMuted.G, colorMuted.B)
	pdf.SetX(marginX)
	subt := fmt.Sprintf("%s · %s · %s", capitalizePtr(lote.Variedad), capitalizePtr(lote.TipoProceso), textPtr(lote.Ubicacion))
	pdf.CellFormat(0, 6, b.tr(subt), "", 1, "L", false, 0, "")
	pdf.Ln(2)
}

func (b *pdfBuilder) drawLoteInfoCard() {
	pdf := b.pdf
	lote := b.data.Lote
	y0 := pdf.GetY()
	cardH := 30.0

	pdf.SetDrawColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
	pdf.SetFillColor(colorRowAlt.R, colorRowAlt.G, colorRowAlt.B)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(marginX, y0, pageW-2*marginX, cardH, 2, "1234", "FD")

	colW := (pageW - 2*marginX - 10) / 2
	rows := [][2]string{
		{"Usuario", b.data.UsuarioNombre},
		{"Peso", formatPesoPtr(lote.PesoKg, " kg")},
		{"Inicio de secado", formatFechaPtr(lote.FechaInicioSecado, "Sin iniciar")},
		{"Fin de secado", formatFechaPtr(lote.FechaFinSecado, "En curso")},
		{"Código QR", lote.CodigoQR},
		{"Tipo de reporte", capitalize(b.data.TipoReporte)},
	}

	x0 := marginX + 5
	y := y0 + 4
	for i, r := range rows {
		col := i % 2
		row := i / 2
		cx := x0 + float64(col)*colW
		cy := y + float64(row)*8.2

		pdf.SetXY(cx, cy)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(colorMuted.R, colorMuted.G, colorMuted.B)
		pdf.CellFormat(colW-5, 4, b.tr(r[0]), "", 0, "L", false, 0, "")

		pdf.SetXY(cx, cy+4)
		pdf.SetFont("Arial", "B", 9.5)
		pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
		pdf.CellFormat(colW-5, 5, b.tr(r[1]), "", 0, "L", false, 0, "")
	}
	pdf.SetY(y0 + cardH + 8)
}

func (b *pdfBuilder) drawKPIRow() {
	pdf := b.pdf
	st := b.data.Estadisticas
	type kpi struct{ label, value string }
	kpis := []kpi{
		{"Temp. promedio", fmt.Sprintf("%.1f °C", st.TemperaturaPromedio)},
		{"Humedad promedio", fmt.Sprintf("%.1f %%", st.HumedadPromedio)},
		{"Días de secado", fmt.Sprintf("%d", st.DiasSecado)},
		{"Lecturas totales", fmt.Sprintf("%d", st.TotalLecturas)},
	}

	gap := 5.0
	cardW := (pageW - 2*marginX - gap*3) / 4
	cardH := 22.0
	b.ensureSpace(cardH + 8)
	y0 := pdf.GetY()

	for i, k := range kpis {
		cx := marginX + float64(i)*(cardW+gap)
		pdf.SetDrawColor(colorGold.R, colorGold.G, colorGold.B)
		pdf.SetFillColor(colorBgDark.R, colorBgDark.G, colorBgDark.B)
		pdf.SetLineWidth(0.3)
		pdf.RoundedRect(cx, y0, cardW, cardH, 2, "1234", "FD")

		pdf.SetXY(cx+3, y0+4)
		pdf.SetFont("Arial", "B", 13)
		pdf.SetTextColor(colorGold.R, colorGold.G, colorGold.B)
		pdf.CellFormat(cardW-6, 7, b.tr(k.value), "", 0, "L", false, 0, "")

		pdf.SetXY(cx+3, y0+13)
		pdf.SetFont("Arial", "", 7.5)
		pdf.SetTextColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
		pdf.CellFormat(cardW-6, 5, b.tr(k.label), "", 0, "L", false, 0, "")
	}
	pdf.SetY(y0 + cardH + 9)
}

func (b *pdfBuilder) sectionTitle(title string) {
	pdf := b.pdf
	b.ensureSpace(16)
	y := pdf.GetY()
	pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
	pdf.Rect(marginX, y+1, 3, 5, "F")
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
	pdf.SetXY(marginX+6, y)
	pdf.CellFormat(0, 7, b.tr(title), "", 1, "L", false, 0, "")
	pdf.SetDrawColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
	pdf.SetLineWidth(0.2)
	pdf.Line(marginX, pdf.GetY()+1, pageW-marginX, pdf.GetY()+1)
	pdf.SetY(pdf.GetY() + 5)
}

type tableCol struct {
	Header string
	Width  float64
	Align  string
}

func (b *pdfBuilder) drawTable(cols []tableCol, rows [][]string, rowFill func(i int) *rgb) {
	pdf := b.pdf
	rowH := 6.5

	drawHead := func() {
		pdf.SetFont("Arial", "B", 8.5)
		pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
		pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
		x := marginX
		y := pdf.GetY()
		for _, c := range cols {
			pdf.SetXY(x, y)
			pdf.CellFormat(c.Width, 7, b.tr(c.Header), "1", 0, c.Align, true, 0, "")
			x += c.Width
		}
		pdf.SetY(y + 7)
	}

	b.ensureSpace(7 + rowH)
	drawHead()

	pdf.SetFont("Arial", "", 8)
	pdf.SetDrawColor(colorGoldSoft.R, colorGoldSoft.G, colorGoldSoft.B)
	pdf.SetLineWidth(0.1)
	for i, row := range rows {
		if pdf.GetY()+rowH > pageBreakTriggerY(pdf) {
			pdf.AddPage()
			drawHead()
			pdf.SetFont("Arial", "", 8)
		}
		fill := colorRowAlt
		if rowFill != nil {
			if c := rowFill(i); c != nil {
				fill = *c
			} else if i%2 == 1 {
				fill = colorWhite
			}
		} else if i%2 == 1 {
			fill = colorWhite
		}
		pdf.SetFillColor(fill.R, fill.G, fill.B)
		pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
		x := marginX
		y := pdf.GetY()
		for ci, c := range cols {
			val := ""
			if ci < len(row) {
				val = row[ci]
			}
			pdf.SetXY(x, y)
			pdf.CellFormat(c.Width, rowH, b.tr(val), "1", 0, c.Align, true, 0, "")
			x += c.Width
		}
		pdf.SetY(y + rowH)
	}
	pdf.SetY(pdf.GetY() + 8)
}

func (b *pdfBuilder) drawLecturasSection() {
	if len(b.data.Lecturas) == 0 {
		return
	}
	b.sectionTitle("Lecturas ambientales")
	cols := []tableCol{
		{"Fecha y hora", 60, "L"},
		{"Temperatura (°C)", 60, "R"},
		{"Humedad (%)", 60, "R"},
	}
	rows := make([][]string, 0, len(b.data.Lecturas))
	for _, l := range b.data.Lecturas {
		rows = append(rows, []string{
			l.Timestamp.Format("02/01/2006 15:04"),
			fmt.Sprintf("%.1f", l.Temperatura),
			fmt.Sprintf("%.1f", l.Humedad),
		})
	}
	b.drawTable(cols, rows, nil)
}

func (b *pdfBuilder) drawAlertasSection() {
	if len(b.data.Alertas) == 0 {
		return
	}
	b.sectionTitle("Alertas")
	cols := []tableCol{
		{"Fecha", 32, "L"},
		{"Severidad", 24, "C"},
		{"Tipo", 40, "L"},
		{"Mensaje", 64, "L"},
		{"Atendida", 20, "C"},
	}
	rows := make([][]string, 0, len(b.data.Alertas))
	for _, a := range b.data.Alertas {
		atendida := "No"
		if a.Atendida {
			atendida = "Sí"
		}
		rows = append(rows, []string{
			a.FechaGenerada.Format("02/01/2006 15:04"),
			strings.ToUpper(a.NivelSeveridad),
			capitalize(strings.ReplaceAll(a.TipoAlerta, "_", " ")),
			truncate(a.Mensaje, 55),
			atendida,
		})
	}
	b.drawTable(cols, rows, func(i int) *rgb {
		switch b.data.Alertas[i].NivelSeveridad {
		case "critica":
			c := rgb{248, 221, 216}
			return &c
		case "alta", "media":
			c := rgb{250, 235, 210}
			return &c
		}
		return nil
	})
}

func (b *pdfBuilder) drawPrediccionesSection() {
	if len(b.data.Predicciones) == 0 {
		return
	}
	b.sectionTitle("Predicciones del modelo")
	cols := []tableCol{
		{"Fecha", 45, "L"},
		{"Tiempo estimado", 45, "R"},
		{"Calidad estimada", 45, "L"},
		{"Confianza", 45, "R"},
	}
	rows := make([][]string, 0, len(b.data.Predicciones))
	for _, p := range b.data.Predicciones {
		tiempoStr := "Pendiente"
		if p.TiempoEstimadoHoras != nil {
			tiempoStr = fmt.Sprintf("%.1f h", *p.TiempoEstimadoHoras)
		}
		confStr := "Pendiente"
		if p.Confianza != nil {
			conf := *p.Confianza
			if conf <= 1.0 {
				conf *= 100
			}
			confStr = fmt.Sprintf("%.0f%%", conf)
		}
		rows = append(rows, []string{
			p.FechaPrediccion.Format("02/01/2006 15:04"),
			tiempoStr,
			capitalizePtr(p.CalidadEstimada),
			confStr,
		})
	}
	b.drawTable(cols, rows, nil)
}

func (b *pdfBuilder) drawRecomendacionesSection() {
	if len(b.data.Recomendaciones) == 0 {
		return
	}
	b.sectionTitle("Recomendaciones")
	pdf := b.pdf
	pdf.SetFont("Arial", "", 9.5)
	for _, r := range b.data.Recomendaciones {
		b.ensureSpace(9)
		y := pdf.GetY()
		pdf.SetFillColor(colorGold.R, colorGold.G, colorGold.B)
		pdf.Circle(marginX+1.5, y+2.7, 1.2, "F")
		pdf.SetTextColor(colorTextDark.R, colorTextDark.G, colorTextDark.B)
		pdf.SetXY(marginX+6, y)
		pdf.MultiCell(pageW-2*marginX-6, 5, b.tr(r.Texto), "", "L", false)
		pdf.Ln(2)
	}
	pdf.Ln(3)
}

func (b *pdfBuilder) drawHistorialSection() {
	if len(b.data.Historial) == 0 {
		return
	}
	b.sectionTitle("Historial de eventos")
	cols := []tableCol{
		{"Fecha", 38, "L"},
		{"Evento", 50, "L"},
		{"Descripción", 92, "L"},
	}
	rows := make([][]string, 0, len(b.data.Historial))
	for _, h := range b.data.Historial {
		rows = append(rows, []string{
			h.FechaEvento.Format("02/01/2006 15:04"),
			capitalize(strings.ReplaceAll(h.TipoEvento, "_", " ")),
			truncate(h.Descripcion, 70),
		})
	}
	b.drawTable(cols, rows, nil)
}
