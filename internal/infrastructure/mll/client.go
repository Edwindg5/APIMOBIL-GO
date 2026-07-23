// Package mll es el cliente HTTP hacia microservicioMLL (el microservicio de Machine
// Learning/PLN de kajve, proyecto Python aparte). Tiene dos llamadas: reportar el tiempo real de
// secado al finalizar un lote, y reportar el puntaje real de catación (escala SCA 0-100) cuando
// exista -- normalmente semanas después, por eso son dos llamadas separadas en el tiempo y no una
// sola. Ambas alimentan retroalimentacion_ml, que es lo que le hace falta al microservicio para
// poder entrenar en el futuro sus modelos de tiempo restante y calidad estimada.
package mll

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client se usa de dos formas distintas a propósito:
//   - ReportarResultadoReal es best-effort: nunca bloquea ni tumba el flujo que lo llama. Si
//     microservicioMLL no responde, no está configurado, o regresa un error, esto solo se
//     registra en el log. Es un extra para reentrenamiento futuro, no una operación crítica -- el
//     lote ya quedó finalizado en la base de datos de este servicio de todas formas.
//   - ReportarCatacion SÍ regresa error al caller: reportar la catación es la única razón de ser
//     de esa llamada (a diferencia de finalizar un lote, no tiene un "efecto principal" del cual
//     ser secundaria), así que si falla, quien la llamó debe enterarse.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient construye el cliente. baseURL vacío (MLL_BASE_URL sin configurar) deja
// ReportarResultadoReal en modo no-op (solo loguea y regresa) y ReportarCatacion regresando un
// error explícito -- igual que el resto de integraciones opcionales del proyecto (ej. FCM del
// lado de microservicioMLL cuando FCM_ENABLED=false).
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) doPost(ctx context.Context, path string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error serializando payload: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Internal-Api-Key", c.apiKey)
	}

	return c.http.Do(req)
}

func (c *Client) doGet(ctx context.Context, path string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-Internal-Api-Key", c.apiKey)
	}

	return c.http.Do(req)
}

type resultadoRealPayload struct {
	TiempoRealHoras *float64 `json:"tiempo_real_horas,omitempty"`
}

// ReportarResultadoReal llama POST /api/v1/internal/lotes/{loteID}/resultado-real en
// microservicioMLL, con el tiempo real de secado (lo único que se conoce al finalizar un lote).
// No regresa error al caller a propósito -- ver comentario del struct Client.
func (c *Client) ReportarResultadoReal(ctx context.Context, loteID int, tiempoRealHoras *float64) {
	if c.baseURL == "" {
		log.Printf("[mll] MLL_BASE_URL no configurado -- se omite resultado-real del lote %d", loteID)
		return
	}

	path := fmt.Sprintf("/api/v1/internal/lotes/%d/resultado-real", loteID)
	resp, err := c.doPost(ctx, path, resultadoRealPayload{TiempoRealHoras: tiempoRealHoras})
	if err != nil {
		log.Printf("[mll] error llamando resultado-real del lote %d: %v", loteID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[mll] resultado-real del lote %d respondió status %d (no crítico: el lote ya quedó finalizado)", loteID, resp.StatusCode)
		return
	}
	log.Printf("[mll] resultado-real del lote %d reportado correctamente a microservicioMLL", loteID)
}

type catacionPayload struct {
	PuntajeSCA float64 `json:"puntaje_sca"`
}

// ReportarCatacion llama POST /api/v1/internal/lotes/{loteID}/catacion en microservicioMLL, con
// el puntaje real de catación (escala SCA 0-100). A diferencia de ReportarResultadoReal, SÍ
// regresa error -- ver comentario del struct Client.
func (c *Client) ReportarCatacion(ctx context.Context, loteID int, puntajeSCA float64) error {
	if c.baseURL == "" {
		return fmt.Errorf("MLL_BASE_URL no configurado, no se puede reportar la catación del lote %d", loteID)
	}

	path := fmt.Sprintf("/api/v1/internal/lotes/%d/catacion", loteID)
	resp, err := c.doPost(ctx, path, catacionPayload{PuntajeSCA: puntajeSCA})
	if err != nil {
		return fmt.Errorf("error llamando catación del lote %d: %w", loteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		detalle, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("catación del lote %d respondió status %d: %s", loteID, resp.StatusCode, string(detalle))
	}
	log.Printf("[mll] catación del lote %d reportada correctamente a microservicioMLL", loteID)
	return nil
}

// ReporteNLG es la respuesta de GET /api/v1/anomalies/{id_lote}/reporte en microservicioMLL: un
// reporte en lenguaje natural (NLG, NLP/generar_reporte.py) que combina alertas, predicciones y
// recomendaciones del lote en un solo texto legible. Se genera al momento en cada llamada
// (siempre refleja el estado actual), no es un texto fijo guardado una sola vez.
type ReporteNLG struct {
	IDReporte    int    `json:"id_reporte"`
	IDLote       int    `json:"id_lote"`
	ReporteTexto string `json:"reporte_texto"`
	FechaGenerado string `json:"fecha_generado"`
}

// ObtenerReporteNLG llama GET /api/v1/anomalies/{loteID}/reporte?id_usuario={usuarioID} en
// microservicioMLL. A diferencia de ReportarResultadoReal/ReportarCatacion, esta SÍ es una
// lectura (no escribe nada del lado de Go), pero igual requiere la API key interna porque
// microservicioMLL protege todo su historial (anomalies) detrás de verificar_api_key -- por eso
// la app móvil no puede llamarla directamente con mlBaseUrl, tiene que pasar por este proxy.
func (c *Client) ObtenerReporteNLG(ctx context.Context, loteID, usuarioID int) (*ReporteNLG, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("MLL_BASE_URL no configurado, no se puede obtener el reporte narrativo del lote %d", loteID)
	}

	path := fmt.Sprintf("/api/v1/anomalies/%d/reporte?id_usuario=%d", loteID, usuarioID)
	resp, err := c.doGet(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error llamando reporte narrativo del lote %d: %w", loteID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lote not found")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("unauthorized")
	}
	if resp.StatusCode >= 300 {
		detalle, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reporte narrativo del lote %d respondió status %d: %s", loteID, resp.StatusCode, string(detalle))
	}

	var reporte ReporteNLG
	if err := json.NewDecoder(resp.Body).Decode(&reporte); err != nil {
		return nil, fmt.Errorf("error decodificando reporte narrativo del lote %d: %w", loteID, err)
	}
	return &reporte, nil
}
