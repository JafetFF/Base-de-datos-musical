package main

import (
       "database/sql"
       "encoding/json"
       "errors"
       "log"
       "net/http"
       _ "modernc.org/sqlite"
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