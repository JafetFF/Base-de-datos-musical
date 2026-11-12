package main

import (
       "database/sql"
       "encoding/json"
       "errors"
       "log"
       "net/http"
       _ "modernc.org/sqlite"
       "Song/Modelo"
)

// Main abre la BDD e inicia el servidor http
func main() {
     db, err := sql.Open("sqlite", "BDD/BaseMusical.db")
     if err != nil {
     	log.Fatal(err)
     }
     defer db.Close()

     if err := db.Ping(); err != nil {
     	log.Fatal(err)
     }

     mux:= http.NewServeMux()

     mux.HandleFunc("GET /api/health", healthHandler(db))
     mux.HandleFunc("GET /api/rolas", rolasHandler(db))

     server := &http.Server{
     	    Addr: "127.0.0.1:8080",
	    Handler: mux,
     }

     log.Println("Servidor inciado en http://127.0.0.1:8080")

     if err := server.ListenAndServe(); err != nil &&
     	!errors.Is(err, http.ErrServerClosed) {
	log.Fatal(err)
     }
}

// healthHandler prueba la conexion con la BDD y regresa su estado
func healthHandler(db *sql.DB) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
     	    if err := db.Ping(); err != nil {
	       http.Error(w, "Error en la BDD", http.StatusServiceUnavailable)
	       return
	    }

	    w.Header().Set("Content-Type", "application/json")
	    json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"database": "connected",
		})
     }
}

// rolasHandler obtiene las rolas de la BDD
// y devuelve sus datos en json
func rolasHandler(db *sql.DB) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
     	    busqueda := r.URL.Query().Get("q")

	    rolas, err := modelo.Busca(busqueda, db)
	    if err != nil {
	       http.Error(w, "Error al consultar las canciones",
	       http.StatusInternalServerError)
	       return
	    }

	    datos := make([]modelo.DatosRola, 0, len(rolas))

	    for _, rola := range rolas {
		datos = append(datos, rola.Datos())
	    }

	    w.Header().Set("Content-Type", "application/json")
	    json.NewEncoder(w).Encode(datos)
     }
}