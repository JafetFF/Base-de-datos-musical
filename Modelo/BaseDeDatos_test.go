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

// Prueba unitaria para el método Agrega
func TestAgrega(t *testing.T) {
     db := bddPrueba(t)

     rola := Rola{
          artist:    "Hillsong United",
          album:     "Unidos Permanecemos",
          albumPath: "/Musica/Hillsong United/Unidos Permanecemos",
          title:     "La Revolución",
          year:      2006,
          genre:     "Pop Rock",
     }

     if err := Agrega(rola, db); err != nil {
         t.Fatalf("No se pudo insertar la rola de prueba: %v", err)
     }

     var titulo string
     var año int
     var genero string
     var artista string
     var album string

     err := db.QueryRow(
         `SELECT r.title, r.year, r.genre, p.name, a.name
          FROM rolas r
          JOIN performers p ON r.id_performer = p.id_performer
          JOIN albums a ON r.id_album = a.id_album
          WHERE r.title = ?`,
          rola.title,
     ).Scan(&titulo, &año, &genero, &artista, &album)

     if err != nil {
         t.Fatalf("Error al encontrar la rola: %v", err)
     }

     if titulo != rola.title ||
         año != rola.year ||
         genero != rola.genre ||
         artista != rola.artist ||
         album != rola.album {
         t.Fatal("No coinciden los datos guardados con los esperados")
     }
}

/*
Función que comprueba que al agregar una rola sin
album se agrega con id_album = 0
*/
func TestAgregaSinAlbum(t *testing.T) {
     db := bddPrueba(t)

     rola := Rola{
         artist: "Eiro Nareth",
         title:  "Around the world",
     }

     if err := Agrega(rola, db); err != nil {
         t.Fatalf("No se pudo insertar la rola: %v", err)
     }

     var idAlbum int
     err := db.QueryRow(
         `SELECT id_album FROM rolas WHERE title = ?`,
         rola.title,
     ).Scan(&idAlbum)

     if err != nil {
         t.Fatal(err)
     }
     if idAlbum != 0 {
         t.Fatalf("Se esperaba id_album = 0, se obtuvo %d", idAlbum)
     }
}

/*
Prueba unitaria que comprueba que no se permitan
artistas repetidos al agregar varias rolas de un mismo artista
*/
func TestAgregaNoDuplicaPerformer(t *testing.T) {
     db := bddPrueba(t)

     rola1 := Rola{
         artist: "Skillet",
         title:  "Hero",
     }

     rola2 := Rola{
         artist: "Skillet",
         title:  "Awake And Alive",
     }

     if err := Agrega(rola1, db); err != nil {
         t.Fatal(err)
     }

     if err := Agrega(rola2, db); err != nil {
         t.Fatal(err)
     }

     var cantidad int

     err := db.QueryRow(
         `SELECT COUNT(*) FROM performers
          WHERE name = ? AND id_type = ?`,
         "Skillet", 2,
     ).Scan(&cantidad)

     if err != nil {
         t.Fatal(err)
     }

     if cantidad != 1 {
         t.Fatalf("Se esperaba 1 performer, se encontraron %d", cantidad)
     }
}

/*
Prueba unitaria que comprueba que no se permitan
álbumes repetidos al agregar varias rolas de un mismo álbum
*/
func TestAgregaNoDuplicaAlbum(t *testing.T) {
     db := bddPrueba(t)

     rola1 := Rola{
         artist:    "Edgar Lira",
         album:     "Morir a Mí (En Vivo)",
         albumPath: "/Musica/Edgar Lira/Morir a Mí (En Vivo)",
         title:     "Entre el Cielo y el Fuego",
     }
     rola2 := Rola{
         artist:    "Edgar lira",
         album:     "Morir a Mí (En Vivo)",
         albumPath: "/Musica/Edgar Lira/Morir a Mí (En Vivo)",
         title:     "Déjate Amar",
     }

     if err := Agrega(rola1, db); err != nil {
         t.Fatal(err)
     }
     if err := Agrega(rola2, db); err != nil {
         t.Fatal(err)
     }

     var cantidad int
     err := db.QueryRow(
         `SELECT COUNT(*) FROM albums
          WHERE name = ? AND path = ?`,
          rola1.album, rola1.albumPath,
     ).Scan(&cantidad)

     if err != nil {
         t.Fatal(err)
     }

     if cantidad != 1 {
         t.Fatalf("Se esperaba 1 álbum, pero se encontraron %d", cantidad)
     }
}

/*
Prueba Unitaria para comprobar que realmente
sirve el rollback del método agrega
*/
func TestAgrega_Rollback(t *testing.T) {
     db := bddPrueba(t)

     _, err := db.Exec(`
         CREATE TRIGGER falla_insercion_rola
         BEFORE INSERT ON rolas
         BEGIN
             SELECT RAISE(ABORT, 'error a propósito');
         END;
     `)
     if err != nil {
         t.Fatal(err)
     }

     rola := Rola{
         artist:    "Hillsong",
         album:     "Hillsong Young & Free - Topic",
         albumPath: "/Musica/Hillsong/Hillsong Young & Free - Topic",
         title:     "Vida Tú Me Das",
         year:      2007,
         genre:     "Pop Rock",
     }

     err = Agrega(rola, db)
     if err == nil {
         t.Fatal("Se esperaba un error al agregar la rola")
     }

     var artistas int
     err = db.QueryRow(`
         SELECT COUNT(*)
         FROM performers
         WHERE name = 'Hillsong'
     `).Scan(&artistas)
     if err != nil {
         t.Fatal(err)
     }

     if artistas != 0 {
         t.Errorf("Se esperaba 0 artistas, se encontraron %d", artistas)
     }

     var albumes int
     err = db.QueryRow(`
         SELECT COUNT(*)
         FROM albums
         WHERE name = 'Hillsong Young & Free - Topic'
     `).Scan(&albumes)
     if err != nil {
         t.Fatal(err)
     }

     if albumes != 0 {
         t.Errorf("Se esperaba 0 álbumes pero se encontraron %d", albumes)
     }
}

/*
Prueba unitaria para comprobar que una rola
sin artista se agrega con artista "Unknown"
*/ 
func TestAgregaSinArtista(t *testing.T) {
     db := bddPrueba(t)

     rola := Rola{
         title: "Desde Mi Interior",
         album: "Con Todo",
         albumPath: "/Musica/Desde Mi Interior/Con Todo",
         year: 2007,
         genre: "Pop Rock",
     }

     err := Agrega(rola, db)
     if err != nil {
         t.Fatal(err)
     }

     var cantidad int
     err = db.QueryRow(`
         SELECT COUNT(*)
         FROM performers
         WHERE name = 'Unknown' AND id_type = 2
     `).Scan(&cantidad)

     if err != nil {
         t.Fatal(err)
     }
     if cantidad != 1 {
         t.Errorf("Se esperaba 1 performer Unknown, se encontraron %d", cantidad)
     }
}

/*
Prueba unitaria que comprueba que dos rolas sin
artistas reutilicen el mismo performer
*/
func TestAgregaNoDuplicaUnknown(t *testing.T) {
     db := bddPrueba(t)

     rolas := []Rola{
         {
             title: "Wait For You",
             year: 2020,
         },
         {
             title: "Hay Otra Salida",
             year: 2007,
         },
     }
     for _, rola := range rolas {
         err := Agrega(rola, db)
         if err != nil {
             t.Fatal(err)
         }
     }

     var cantidad int
     err := db.QueryRow(`
         SELECT COUNT(*)
         FROM performers
         WHERE name = 'Unknown' AND id_type = 2
     `).Scan(&cantidad)

     if err != nil {
         t.Fatal(err)
     }
     if cantidad != 1 {
         t.Errorf("Se esperaba 1 performer Unknown, pero se encontraron %d", cantidad)
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
