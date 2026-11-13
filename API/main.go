package main

import (
       "database/sql"
       "encoding/json"
       "errors"
       "log"
       "net/http"
       _ "modernc.org/sqlite"
       "Song/Modelo"
       "strings"
       "os"
       "path/filepath"
       "strconv"
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
     mux.HandleFunc("POST /api/minar", minarHandler(db))
     mux.HandleFunc("GET /api/portadas/{id}", portadaHandler)
     mux.HandleFunc("GET /api/personas", personasHandler(db))
     mux.HandleFunc("POST /api/personas", agregaPersonaHandler(db)) 

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

	    busqueda = strings.TrimSpace(busqueda)
	    if busqueda == "" {
	       busqueda = `title~""`
	    } else if !strings.ContainsAny(busqueda, "=<>!~&|") {
	      busqueda = "title~" + strconv.Quote(busqueda)
	    }
	    rolas, err := modelo.BuscaCompilada(busqueda, db)
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

// minarHandler recibe un directorio y ejecuta al minero
func minarHandler(db *sql.DB) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
            var solicitud struct {
            	Directorio string `json:"directorio"`
            }

            err := json.NewDecoder(r.Body).Decode(&solicitud)
            if err != nil {
               http.Error(w, "Solicitud de json inválida",
               http.StatusBadRequest)
               return
            }

            if strings.TrimSpace(solicitud.Directorio) == "" {
               http.Error(w, "Debes indicar un directorio a minar",
               http.StatusBadRequest)
               return
            }

            err = modelo.MinaYAgrega(solicitud.Directorio, db)
            if err != nil {
               http.Error(w, "Ocurrió un error al minar el directorio: "+err.Error(),
               http.StatusInternalServerError)
               return
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
            	"message": "Se minó correctamente",
            })
     }
}

// portadaHandler regresa la portada de un álbum
func portadaHandler(w http.ResponseWriter, r *http.Request) {
     id, err := strconv.Atoi(r.PathValue("id"))
     if err != nil || id <= 0 {
        http.Error(w, "ID álbum inválido", http.StatusBadRequest)
        return
     }

     extensiones := []string{".jpg", ".png", ".gif", ".webp"}

     for _, extension := range extensiones {
         ruta := filepath.Join(
              "BDD",
              "Portadas",
              strconv.Itoa(id)+extension,
         )

         imagen, err := os.ReadFile(ruta)
         if err != nil {
            continue
         }

         w.Header().Set("Content-Type", http.DetectContentType(imagen))
         w.Write(imagen)
         return
     }

     http.Error(w, "No se encontró la portada", http.StatusNotFound)
}

// personasHandler obtiene todas las personas de la BDD
// y devuelve sus datos en json
func personasHandler(db *sql.DB) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
            personas, err := modelo.BuscaPersonas(db)
            if err != nil {
               http.Error(w, "Ocurrió un error al consultar las personas",
               http.StatusInternalServerError)
               return
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(personas)
     }
}

// agregaPersonaHandler registra a una persona en la BDD
func agregaPersonaHandler(db *sql.DB) http.HandlerFunc {
     return func(w http.ResponseWriter, r *http.Request) {
            var persona modelo.Persona
            err := json.NewDecoder(r.Body).Decode(&persona)
            if err != nil {
               http.Error(w, "Solicitud de json inválida",
               http.StatusBadRequest)
               return
            }

            if strings.TrimSpace(persona.NombreArtístico) == "" &&
	       strings.TrimSpace(persona.NombreReal) == "" {
               http.Error(w, "Proporciona al menos un nombre",
               http.StatusBadRequest)
               return
            }

            err = modelo.AgregaPersona(persona, db)
            if err != nil {
               http.Error(w, "Error al registrar la persona",
               http.StatusInternalServerError)
               return
            }

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)

            json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
            	"message": "Persona registrada correctamente",
            })
    }
}