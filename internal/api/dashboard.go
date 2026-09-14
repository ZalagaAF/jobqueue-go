package api

import (
	_ "embed"
	"net/http"
)

// dashboardHTML queda embebido en el binario compilado gracias a
// //go:embed — no depende de que el archivo .html exista suelto en
// el filesystem del servidor en producción, viaja adentro del
// ejecutable. La directiva debe ir pegada, sin línea en blanco,
// justo arriba de la variable que va a contener el contenido.
//
//go:embed dashboard.html
var dashboardHTML []byte

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML)
}