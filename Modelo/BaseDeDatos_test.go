package main

import (
       "database/sql"
       "os"
       "path/filepath"
       "testing"
       _ "modernc.org/sqlite"
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

// Prueba unitaria para el método Busca
func TestBusca(t *testing.T) {
     db := bddPrueba(t)

     _, err := db.Exec(
     	`INSERT INTO rolas (title, year, genre)
	VALUES (?, ?, ?)`, "RolaPrueba", 2026, "Rock")

     if err != nil {
	t.Fatalf("No se pudo insertar la rola de prueba: %v", err)
     }

     rolas, err := Busca("RolaPrueba", db)
     if err != nil {
	t.Fatalf("La búsqueda devolvió un error: %v", err)
     }
     
     if len(rolas) == 0 {
	t.Fatal("Se esperaba encontrar la rola pero no hubo resultados")
     }
     
     encontrada := false

     for _, rola := range rolas {
          if rola.title == "RolaPrueba" {
	     encontrada = true
	     break
	  }
     }

     if !encontrada {
	t.Fatal("Se encontraron resultados pero no la rola esperada")
     }
}

// Prueba unitaria para el método Busca
func TestAgrega(t *testing.T) {    
     db := bddPrueba(t)

     rola := Rola{
     	  title: "La Revolución",
	  year:  2006,
	  genre: "Pop Rock",
     }
     
     if err := Agrega(rola, db); err != nil {
	t.Fatalf("No se pudo insertar la rola de prueba: %v", err)
     }

     var titulo string
     var año int
     var genero string
     
     err := db.QueryRow(
     	 `SELECT title, year, genre FROM rolas
	 WHERE title = ?`, rola.title).Scan(&titulo, &año, &genero)

     if err != nil {
     	t.Fatalf("Error al encontrar la rola: %v", err)
     }

     if titulo != rola.title ||
        año != rola.year ||
	genero != rola.genre {
     	t.Fatal("No coinciden los datos guardados con los esperados")
     }
}

// Prueba unitaria para el método Elimina
func TestElimina(t *testing.T) {    
     db := bddPrueba(t)

     result, err := db.Exec(
     	     `INSERT INTO rolas (title) VALUES (?)`, "Prueba")

     if err != nil {
     	t.Fatalf("No se añadió la rola: %v", err)
     }

     id, err := result.LastInsertId()
     if err != nil {
     	t.Fatalf("No se pudo obtener el id: %v", err)
     }

     if err := Elimina(int(id), db); err != nil {
     	t.Fatalf("Ocurrió un problema al eliminar la rola: %v", err)
     }

     var ocurrencias int
     err = db.QueryRow(
     	 `SELECT COUNT(*) FROM rolas
	 WHERE id_rola = ?`, id).Scan(&ocurrencias)

     if err != nil {
     	t.Fatalf("Ocurrió un problema al buscar coincidencias: %v", err)
     }

     if ocurrencias != 0 {
     	t.Fatal("La rola no fue eliminada")
     }
}

// Prueba unitaria para el método Edita
func TestEdita(t *testing.T) {    
     db := bddPrueba(t)

     original := Rola{
     	      title: "Quizá",
	      track: 2,
	      year:  2010,
	      genre: "Balada pop",
     }

     if err := Agrega(original, db); err != nil {
     	t.Fatalf("Error al agregar la rola: %v", err)
     }

     var id int
     err := db.QueryRow(
     	 `SELECT id_rola FROM rolas
	  WHERE title = ?`, original.title).Scan(&id)

     if err != nil {
     	t.Fatal(err)
     }

     modificada := Rola{
     		title: "What I've Done",
		track: 6,
		year:  2007,
		genre: "Hard Rock",
     }

     if err := Edita(id, modificada, db); err != nil {
     	t.Fatalf("Error al editar la rola: %v", err)
     }

     var res Rola
     err = db.QueryRow(
     	 `SELECT id_rola, title, track, year, genre
	 FROM rolas
	 WHERE id_rola = ?`, id).Scan(
	 &res.id, &res.title,
	 &res.track, &res.year, &res.genre,
         )

     if err != nil {
     	t.Fatal(err)
     }

     if res.id != id {
     	t.Fatalf("Cambió el ID de %d a %d", id, res.id)
     }

     if res.title != modificada.title ||
     	res.track != modificada.track ||
	res.year  != modificada.year  ||
	res.genre != modificada.genre {
	t.Fatal("No se modificaron los datos de la rola correctamente")
     }
}
