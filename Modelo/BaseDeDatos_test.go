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

/*
Prueba Unitaria para verificar que si se
hace una búsqueda y no hay coincidencias,
entonces no se devuelve ni error ni resultado
*/
func TestBuscaSinResultados(t *testing.T) {
     db := bddPrueba(t)

     rolas, err := Busca("Algo", db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 0 {
         t.Fatalf("Se esperaban 0 resultados, se encontraron %d", len(rolas))
     }
}

/*
Prueba Unitaria que comprueba que podemos
encontrar una rola con solo una parte del título
*/
func TestBusquedaParcial(t *testing.T) {
     db := bddPrueba(t)

     _, err := db.Exec(
         `INSERT INTO rolas (title, year, genre)
          VALUES (?, ?, ?)`,
         "Rola específica", 1500, "Indie",
     )
     if err != nil {
         t.Fatal(err)
     }

     rolas, err := Busca("Rola", db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error %v", err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 resultado, pero se encontraron %d", len(rolas))
     }

     if rolas[0].title != "Rola específica" {
         t.Errorf("Se encontró %q en lugar de Rola específica", rolas[0].title)
     }
}

/*
Prueba unitaria que comprueba que podemos
encontrar varios resultados a partir de una
misma búsqueda
*/
func TestBusquedaMultiple(t *testing.T) {
     db := bddPrueba(t)

     titulos := []string{
         "Wait For You",
         "Wait For The Moment",
         "Wait For The Day",
     }

     for _, titulo := range titulos {
         _, err := db.Exec(
             `INSERT INTO rolas (title, year, genre) VALUES (?, ?, ?)`,
             titulo, 2026, "Pop",
         )
         if err != nil {
             t.Fatal(err)
         }
     }

     rolas, err := Busca("Wait", db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error %v", err)
     }

     if len(rolas) != len(titulos) {
         t.Fatalf("Se esperaban %d resultados, se encontraron %d",
             len(titulos), len(rolas))
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
     	t.Errorf("Se esperaba 0 coincidencias pero se encontraron %d", ocurrencias)
     }
}

/*
Prueba Unitaria para comprobar que da error
al intentar eliminar una rola que no existe
*/
func TestEliminaInexistente(t *testing.T) {
     db := bddPrueba(t)

     err := Elimina(777, db)
     if err == nil {
         t.Error("Se esperaba un error al eliminar una rola que no existe")
     }
}

/*
Prueba Unitaria para ver que eliminar una rola
no afecta a las demás rolas de un mismo artista
*/
func TestEliminaNoAfectaOtrasRolas(t *testing.T) {
     db := bddPrueba(t)

     rolas := []Rola{
         {title: "Call To Me", artist: "RIDERS"},
         {title: "Fake Fix", artist: "RIDERS"},
     }

     for _, rola := range rolas {
         err := Agrega(rola, db)
         if err != nil {
             t.Fatal(err)
         }
     }

     var id int
     err := db.QueryRow(
         `SELECT id_rola FROM rolas WHERE title = ?`,
         "Call To Me",
     ).Scan(&id)
     if err != nil {
         t.Fatal(err)
     }

     err = Elimina(id, db)
     if err != nil {
         t.Fatal(err)
     }

     var cantidad int
     err = db.QueryRow(
         `SELECT COUNT(*) FROM rolas WHERE title = ?`,
         "Fake Fix",
     ).Scan(&cantidad)
     if err != nil {
         t.Fatal(err)
     }

     if cantidad != 1 {
         t.Errorf("Se esperaba que quedara 1 rola, y se encontraron %d", cantidad)
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

/*
Prueba Unitaria para comprobar que se arroja un
error al intentar editar una rola inexistente
*/
func TestEditaInexistente(t *testing.T) {
     db := bddPrueba(t)

     modificada := Rola{
         title: "What I've Done",
         track: 6,
         year: 2007,
         genre: "Hard Rock",
     }

     err := Edita(1234, modificada, db)

     if err == nil {
         t.Fatal("Se esperaba un error al editar una rola inexistente")
     }
}

/*
Prueba unitaria para verificar que efectivamente
editar una rola no afecta a las demás rolas
*/
func TestEditaNoModificaOtrasRolas(t *testing.T) {
     db := bddPrueba(t)

     rolas := []Rola{
         {
             title: "Quizá",
             track: 2,
             year: 2010,
             genre: "Balada pop",
         },
         {
             title: "Need Your Love",
             track: 5,
             year: 2003,
             genre: "Pop Rock",
         },
     }

     for _, rola := range rolas {
         if err := Agrega(rola, db); err != nil {
           t.Fatal(err)
         }
     }

     var id int
     err := db.QueryRow(
         `SELECT id_rola FROM rolas WHERE title = ?`,
         "Quizá",
     ).Scan(&id)

     if err != nil {
         t.Fatal(err)
     } 

     modificada := Rola{
         title: "What I've Done",
         track: 6,
         year: 2007,
         genre: "Hard Rock",
     }

     if err := Edita(id, modificada, db); err != nil {
         t.Fatal(err)
     } 

     var res Rola
     err = db.QueryRow(
         `SELECT title, track, year, genre
          FROM rolas
          WHERE title = ?`,
         "Need Your Love",
     ).Scan(&res.title, &res.track, &res.year, &res.genre)

     if err != nil {
         t.Fatal(err)
     }

     if res.title != "Need Your Love" ||
        res.track != 5 ||
        res.year != 2003 ||
        res.genre != "Pop Rock" {
        t.Fatal("Se modificó una rola que no era")
     }
}

/*
Función auxiliar que inserta datos de prueba
en la base de datos temporal
*/
func insertaDatosDeIntegracion(t *testing.T, db *sql.DB) {
     t.Helper()

     _, err := db.Exec(`
         INSERT INTO performers VALUES
         (1, 0, 'Eiro Nareth'),
         (2, 1, 'Bernth')
     `)
     if err != nil {
         t.Fatalf("Error al insertar los artistas: %v", err)
     }

     _, err = db.Exec(`
         INSERT INTO albums VALUES
         (1, '/music/Acoustic', 'Acoustic', 2007),
         (2, '/music/Electric', 'Electric', 2006)
     `)
     if err != nil {
         t.Fatalf("Error al insertar los álbumes: %v", err)
     }

     _, err = db.Exec(`
         INSERT INTO rolas VALUES
         (1, 1, 1, '/music/Acoustic/Rola1.mp3',
          'Acoustic', 1, 2007, 'Rock'),
 
         (2, 2, 2, '/music/Electric/Rola2.mp3',
          'Electric', 6, 2006, 'Progressive Rock'),

         (3, 1, 1, '/music/Imagine/Rola3.mp3',
          'Other', 2, 2004, 'Pop')
     `)
     if err != nil {
         t.Fatalf("Error al insertar las rolas: %v", err)
     }
}

/*
Prueba de integración sobre el lexer, parser,
traductor y la BDD
*/
func TestBuscaCompiladaIntegracion(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(`title = "Acoustic"`, db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 canción, pero se obtuvieron %d",
	 len(rolas))
     }

     if rolas[0].title != "Acoustic" {
         t.Errorf("Se esperaba Acoustic, se obtuvo %s", rolas[0].title)
     }

     if rolas[0].artist != "Eiro Nareth" {
         t.Errorf("Se esperaba Eiro Nareth, pero se obtuvo %s",
	 rolas[0].artist)
     }
}

/*
Prueba de integración que verifica las búsquedas
con múltiples condiciones y operadores lógicos
*/
func TestBuscaCompiladaOperacionLogica(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(
         `artist = "Eiro Nareth" & year = 2007`, db,
     )
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 canción, pero se obtuvieron %d",
         len(rolas))
     }

     if rolas[0].title != "Acoustic" {
         t.Errorf("Se esperaba Acoustic, se obtuvo %s",
         rolas[0].title)
     }
 
     if rolas[0].year != 2007 {
         t.Errorf("Se esperaba el año 2007, se obtuvo %d",
         rolas[0].year)
     }

     if rolas[0].artist != "Eiro Nareth" {
         t.Errorf("Se esperaba Eiro Nareth, se obtuvo %s",
         rolas[0].artist)
     }
}

/*
Prueba de integración que verifica las búsquedas
con el operador lógico | (OR)
*/
func TestBuscaCompiladaOR(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(
         `artist = "Eiro Nareth" | year = 2006`, db,
     )
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 3 {
         t.Fatalf("Se esperaban 3 canciones, pero se obtuvieron %d",
         len(rolas))
     }
}

/*
Prueba de integración que verifica las búsquedas
con el operador lógico ! (NOT)
*/
func TestBuscaCompiladaNOT(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(
         `! (artist = "Eiro Nareth")`, db,
     )
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }
     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 canción, pero se obtuvieron %d",
         len(rolas))
     }

     if rolas[0].title != "Electric" {
         t.Errorf("Se esperaba Electric, se obtuvo %s",
         rolas[0].title)
     }
}

/*
Prueba de integración que verifica que los paréntesis
modifiquen correctamente la precedencia de las operaciones lógicas
*/
func TestBuscaCompiladaPrecedencia(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(
         `(artist = "Eiro Nareth" | artist = "Bernth") & year = 2007`,
         db,
     )
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 canción, pero se obtuvieron %d",
         len(rolas))
     }
     if rolas[0].title != "Acoustic" {
         t.Errorf("Se esperaba Acoustic, se obtuvo %s",
         rolas[0].title)
     }
}

/*
Prueba de integración que verifica que las canciones
con campos NULL puedan obtenersee correctamente
*/
func TestBuscaCompiladaNULL(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     _, err := db.Exec(`
         INSERT INTO rolas VALUES
         (4, 1, 1, NULL, 'Sin datos', NULL, NULL, NULL)
     `)
     if err != nil {
         t.Fatalf("Error al insertar la canción: %v", err)
     }

     rolas, err := BuscaCompilada(`title = "Sin datos"`, db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 canción, pero se obtuvieron %d",
         len(rolas))
     }

     if rolas[0].title != "Sin datos" {
         t.Errorf("Se esperaba Sin datos, se obtuvo %s",
         rolas[0].title)
     }
}

/*
Prueba de integración que verifica que una búsqueda
sin ocurrencias devuelva una colección vacía
*/
func TestBuscaCompiladaSinResultados(t *testing.T) {
     db := bddPrueba(t)
     insertaDatosDeIntegracion(t, db)

     rolas, err := BuscaCompilada(`title = "Inexistente"`, db)
     if err != nil {
         t.Fatalf("La búsqueda devolvió un error: %v", err)
     }

     if len(rolas) != 0 {
         t.Fatalf("Se esperaban 0 canciones, pero se obtuvieron %d",
         len(rolas))
     }
}