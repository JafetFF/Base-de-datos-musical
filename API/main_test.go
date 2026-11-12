package main

import (
	"testing"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

// Función que regresa una bdd de prueba con el esquema
func bddPrueba(t *testing.T) *sql.DB {
     t.Helper()

     ruta := filepath.Join(t.TempDir(), "prueba.db")
     db, err := sql.Open("sqlite", ruta)
     if err != nil {
     	t.Fatalf("No se pudo abrir la base de datos: %v", err)
     }
     t.Cleanup(func() {
     	db.Close()
     })

     if err := db.Ping(); err != nil {
	t.Fatalf("No se pudo conectar a la base de datos: %v", err)
     }

     esquema, err := os.ReadFile("../BDD/rolas.sql")
     if err != nil {
	t.Fatalf("No se pudo leer el esquema: %v", err)
     }
     
     if _, err := db.Exec(string(esquema)); err != nil {
	t.Fatalf("No se pudo ejecutar el esquema: %v", err)
     }

     return db
}

// TestRolasHandler comprueba que se reciba el json correspondiente
func TestRolasHandler(t *testing.T) {
     db := bddPrueba(t)
	
     defer db.Close()

     consultas := []string{
     	       `INSERT INTO performers VALUES (1, 0, 'Metallica')`,
	       `INSERT INTO albums VALUES (1, '/musica/Ride The Lightning (Deluxe Remaster)', 'Ride The Lightning (Deluxe Remaster)', 1984)`,
	       `INSERT INTO rolas VALUES (1, 1, 1, '/musica/Ride The Lightning (Deluxe Remaster)/IMAGINE.mp3',
 	       'Creeping Death', 1, 1984, 'Metal')`,
     }

     for _, consulta := range consultas {
	 _, err := db.Exec(consulta)
	 if err != nil {
	    t.Fatal(err)
	 }
     }

     request := httptest.NewRequest("GET", "/api/rolas", nil)
     recorder := httptest.NewRecorder()

     rolasHandler(db).ServeHTTP(recorder, request)

     if recorder.Code != 200 {
	t.Fatalf("Se esperaba 200 pero se obtuvo %d", recorder.Code)
     }
     var rolas []map[string]interface{}

     err := json.Unmarshal(recorder.Body.Bytes(), &rolas)
     if err != nil {
	t.Fatal(err)
     }

     if len(rolas) != 1 {
	t.Fatalf("Se esperaba 1 rola, y se recibieron %d", len(rolas))
     }

     if rolas[0]["title"] != "Creeping Death" {
	t.Errorf("Se esperaba Creeping Death pero se obtuvo %v", rolas[0]["title"])
     }
}

// TestRolasPruebaSinRolas verifica el caso cuando no hay rolas
func TestRolasHandlerSinRolas(t *testing.T) {
     db := bddPrueba(t)

     request := httptest.NewRequest("GET", "/api/rolas", nil)
     recorder := httptest.NewRecorder()

     rolasHandler(db).ServeHTTP(recorder, request)

     if recorder.Code != 200 {
        t.Fatalf("Se esperaba 200 pero se obtuvo %d", recorder.Code)
     }

     var rolas []map[string]interface{}

     err := json.Unmarshal(recorder.Body.Bytes(), &rolas)
     if err != nil {
        t.Fatal(err)
     }

     if len(rolas) != 0 {
        t.Errorf("Se esperaban 0 rolas, y se recibieron %d", len(rolas))
     }
}