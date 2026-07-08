package reportgen

// Paleta de colores de la app (fondo oscuro café + dorado), usada tanto
// en el PDF (RGB) como en el Excel (hex ARGB) para que el reporte se
// sienta parte de la misma identidad visual.
var (
	colorBgDark    = rgb{26, 18, 15}    // #1A120F fondo café oscuro (portada/encabezados)
	colorCard      = rgb{43, 30, 24}    // #2B1E18 tarjetas sobre fondo oscuro
	colorGold      = rgb{227, 167, 59}  // #E3A73B dorado principal (acento de marca)
	colorGoldSoft  = rgb{240, 214, 160} // #F0D6A0 dorado claro (líneas/bordes suaves)
	colorTextDark  = rgb{43, 26, 10}    // #2B1A0A texto sobre fondo dorado
	colorTextLight = rgb{247, 240, 232} // #F7F0E8 texto claro sobre fondo oscuro
	colorMuted     = rgb{140, 122, 105} // #8C7A69 texto secundario/gris cálido
	colorRowAlt    = rgb{250, 244, 234} // #FAF4EA fila alterna clara en tablas
	colorWhite     = rgb{255, 255, 255}
	colorDanger    = rgb{176, 58, 46} // #B03A2E alertas críticas
	colorWarning   = rgb{196, 130, 40} // #C48228 alertas medias
)

type rgb struct {
	R, G, B int
}

const (
	hexBgDark   = "1A120F"
	hexGold     = "E3A73B"
	hexGoldSoft = "F0D6A0"
	hexTextDark = "2B1A0A"
	hexRowAlt   = "FAF4EA"
	hexWhite    = "FFFFFF"
	hexDanger   = "B03A2E"
	hexWarning  = "C48228"
)
