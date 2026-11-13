package modelo

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

/*
Prueba unitaria para la función BuscaArtistas
*/
func TestBuscaArtistas(t *testing.T) {
     db := bddPrueba(t)
     
     defer db.Close()

     _, err := db.Exec(
        `INSERT INTO performers VALUES
        (1, 0, 'Adele'),
        (2, 1, 'Omar'),
        (3, 2, 'Unknown')
     `)
     if err != nil {
        t.Fatal(err)
     }

     artistas, err := BuscaArtistas(db)
     if err != nil {
        t.Fatalf("Ocurrió un error al buscar artistas: %v", err)
     }

     if len(artistas) != 3 {
        t.Fatalf("Se esperaban 3 artistas, pero se obtuvieron %d", len(artistas))
     }

     esperados := []Artista{
               {Id: 1, Nombre: "Adele", Tipo: 0},
       	       {Id: 2, Nombre: "Omar", Tipo: 1},
               {Id: 3, Nombre: "Unknown", Tipo: 2},
     }

     for i, artista := range artistas {
         if artista != esperados[i] {
            t.Errorf("Artista %d incorrecto. Se esperaba %+v, se obtuvo %+v",
	    i, esperados[i], artista)
         }
     }
}

// Prueba unitaria para AgregaPersona
func TestAgregaPersona(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     persona := Persona{
             NombreArtístico:    "Saleka",
             NombreReal:         "Saleka Shyamalan",
             FechaNacimiento:    "01/08/1996",
             FechaFallecimiento: "",
     }

     err := AgregaPersona(persona, db)
     if err != nil {
        t.Fatalf("Error al agregar persona: %v", err)
     }
     var idPerformer, idPerson int
     var nombre, nombreReal, nacimiento, fallecimiento string

     err = db.QueryRow(
         `SELECT performers.id_performer, persons.id_person,
	 performers.name, persons.real_name,
         persons.birth_date, persons.death_date
         FROM performers
         JOIN persons ON performers.id_performer = persons.id_person
     `).Scan(
	&idPerformer,
        &idPerson,
        &nombre,
        &nombreReal,
        &nacimiento,
        &fallecimiento,
     )
     if err != nil {
        t.Fatalf("Ocurrió un error al consultar los registros: %v", err)
     }
     if idPerformer != idPerson {
        t.Errorf("Los id's no coinciden: %d y %d",
        idPerformer, idPerson)
     }

     if nombre != persona.NombreArtístico {
        t.Errorf("Nombre artístico incorrecto: %s", nombre)
     }

     if nombreReal != persona.NombreReal {
        t.Errorf("El nombre real es incorrecto: %s", nombreReal)
     }

     if nacimiento != persona.FechaNacimiento {
        t.Errorf("Fecha de nacimiento incorrecta: %s", nacimiento)
     }

     if fallecimiento != "" {
        t.Errorf("La fecha de fallecimiento debería estar vacía: %s",
        fallecimiento)
     }
}

// Prueba para la función BuscaPersona
func TestBuscaPersona(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     _, err := db.Exec(`
        INSERT INTO persons VALUES (
        1, 'Kurt Cobain',
        'Kurt Donald Cobain',
        '20/02/1967',
        '05/04/1994'
        );
     `)
     if err != nil {
        t.Fatal(err)
     }

     persona, err := BuscaPersona(1, db)
     if err != nil {
        t.Fatalf("Error al buscar persona: %v", err)
     }

     if persona.ID != 1 {
        t.Errorf("Id incorrecto: %d", persona.ID)
     }

     if persona.NombreArtístico != "Kurt Cobain" {
        t.Errorf("Nombre artístico incorrecto: %s",
        persona.NombreArtístico)
     }

     if persona.NombreReal != "Kurt Donald Cobain" {
        t.Errorf("Nombre real incorrecto: %s",
        persona.NombreReal)
     }

     if persona.FechaNacimiento != "20/02/1967" {
        t.Errorf("Fecha de nacimiento incorrecta: %s",
        persona.FechaNacimiento)
     }

     if persona.FechaFallecimiento != "05/04/1994" {
        t.Errorf("Fecha de fallecimiento incorrecta: %s",
        persona.FechaFallecimiento)
     }
}

/*
Test que verifica que se envía un error
al buscar una persona que no está registrada
*/
func TestBuscaPersonaNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     _, err := BuscaPersona(99, db)
 
     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba un error pero se obtuvo: %v", err)
     }
}

// Prueba para la función EditaPersona
func TestEditaPersona(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     _, err := db.Exec(`
        INSERT INTO performers VALUES (1, 0, 'Kurt Cobain');
        INSERT INTO persons VALUES (
               1, 'Kurt Cobain',
               'Kurt Donald Cobain',
               '20/02/1967',
               '05/04/1994'
        );
     `)
     if err != nil {
        t.Fatal(err)
     }

     persona := Persona{
             ID:                 1,
             NombreArtístico:    "Kurt Cobain (Nirvana)",
             NombreReal:         "Kurt Donald Cobain",
             FechaNacimiento:    "15/09/1960",
             FechaFallecimiento: "05/02/1996",
     }

     err = EditaPersona(persona, db)
     if err != nil {
        t.Fatalf("Error al editar persona: %v", err)
     }

     var nombrePerformer, nombreArtístico, nombreReal string

     err = db.QueryRow(`
         SELECT performers.name,
                persons.stage_name,
                persons.real_name
         FROM performers
         JOIN persons ON performers.id_performer = persons.id_person
         WHERE persons.id_person = 1
     `).Scan(
	&nombrePerformer,
        &nombreArtístico,
        &nombreReal,
     )

     if err != nil {
        t.Fatal(err)
     }

     if nombrePerformer != persona.NombreArtístico {
        t.Errorf("Nombre del intérprete incorrecto: %s", nombrePerformer)
     }

     if nombreArtístico != persona.NombreArtístico {
        t.Errorf("Nombre artístico incorrecto: %s", nombreArtístico)
     }

     if nombreReal != persona.NombreReal {
        t.Errorf("Nombre real incorrecto: %s", nombreReal)
     }
}

/*
Test para verificar que se devuelve un error
al intentar editar una persona inexistente
*/
func TestEditaPersonaNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     persona := Persona{
             ID:              99,
             NombreArtístico: "Persona inexistente",
     }

     err := EditaPersona(persona, db)

     if err == nil {
        t.Error("Se esperaba un error al editar una persona inexistente")
     }
}

/*
Test para verificar que se añade
correctamente un grupo a la BDD
*/
func TestAgregaGrupo(t *testing.T) {
     db := bddPrueba(t)

     defer db.Close()
   
     grupo := Grupo{
           Nombre:      "Metallica",
           FechaInicio: "28/10/1981",
           FechaFin:    "",
     }

     err := AgregaGrupo(grupo, db)
     if err != nil {
        t.Fatalf("Error al agregar grupo: %v", err)
     }

     var idPerformer, idGroup, tipo int
     var nombrePerformer, nombreGrupo, inicio, fin string

     err = db.QueryRow(
         `SELECT performers.id_performer,
                performers.id_type,
                performers.name,
                groups.id_group,
                groups.name,
                groups.start_date,
                groups.end_date
         FROM performers
         JOIN groups ON performers.id_performer = groups.id_group
     `).Scan(
	&idPerformer,
        &tipo,
        &nombrePerformer,
        &idGroup,
        &nombreGrupo,
        &inicio,
        &fin,
     )

     if err != nil {
        t.Fatalf("Error al consultar los registros: %v", err)
     }

     if idPerformer != idGroup {
        t.Errorf("Los id's no coinciden: %d y %d",
        idPerformer, idGroup)
     }
     if tipo != 1 {
        t.Errorf("El tipo debería ser 1 (de Grupo), se obtuvo: %d", tipo)
     }
     if nombrePerformer != grupo.Nombre {
        t.Errorf("Nombre del artista incorrecto: %s", nombrePerformer)
     }
     if nombreGrupo != grupo.Nombre {
        t.Errorf("Nombre del grupo incorrecto: %s", nombreGrupo)
     }
     if inicio != grupo.FechaInicio {
        t.Errorf("Fecha de inicio incorrecta: %s", inicio)
     }
     if fin != grupo.FechaFin {
        t.Errorf("Fecha de final incorrecta: %s", fin)
     }
}

/*
Prueba para verificar la función buscaGrupo
*/
func TestBuscaGrupo(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     _, err := db.Exec(`   
        INSERT INTO groups VALUES (
               1, 'Metallica', '28/10/1981', ''
        );
     `)
     if err != nil {
        t.Fatal(err)
     }

     grupo, err := BuscaGrupo(1, db)
     if err != nil {
        t.Fatalf("Ocurrió un error al buscar grupo: %v", err)
     }
     if grupo.ID != 1 {
        t.Errorf("Id incorrecto: %d", grupo.ID)
     }
     if grupo.Nombre != "Metallica" {
        t.Errorf("Nombre incorrecto: %s", grupo.Nombre)
     }
     if grupo.FechaInicio != "28/10/1981" {
        t.Errorf("Fecha de inicio incorrecta: %s", grupo.FechaInicio)
     }
     if grupo.FechaFin != "" {
        t.Errorf("Fecha de fin incorrecta: %s", grupo.FechaFin)
     }
}

/*
Prueba para verificar que se devuelve un error
al intentar buscar un grupo inexistente
*/
func TestBuscaGrupoNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     _, err := BuscaGrupo(1, db)

     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba un error pero se obtuvo: %v", err)
     }
}

/*
Prueba para EditaGrupo
*/
func TestEditaGrupo(t *testing.T) {
     db := bddPrueba(t)
     
     defer db.Close()

     _, err := db.Exec(`
        INSERT INTO performers VALUES (1, 1, 'Metallica');
        INSERT INTO groups VALUES (1, 'Metallica', '28/10/1981', '');
     `)
     if err != nil {
        t.Fatal(err)
     }

     grupo := Grupo{
           ID:          1,
           Nombre:      "Metalica",
           FechaInicio: "20/09/1989",
           FechaFin:    "14/08/2024",
     }

     err = EditaGrupo(grupo, db)
     if err != nil {
        t.Fatalf("Error al editar grupo: %v", err)
     }

     var nombrePerformer string
     var nombreGrupo, inicio, fin string

     err = db.QueryRow(
         `SELECT name FROM performers WHERE id_performer = 1
     `).Scan(&nombrePerformer)
     if err != nil {
        t.Fatal(err)
     }

     err = db.QueryRow(
         `SELECT name, start_date, end_date
         FROM groups
         WHERE id_group = 1
     `).Scan(&nombreGrupo, &inicio, &fin)
     if err != nil {
        t.Fatal(err)
     }

     if nombrePerformer != grupo.Nombre {
        t.Errorf("Nombre del artista incorrecto: %s", nombrePerformer)
     }

     if nombreGrupo != grupo.Nombre {
        t.Errorf("Nombre del grupo incorrecto: %s", nombreGrupo)
     }

     if inicio != grupo.FechaInicio {
        t.Errorf("Fecha de inicio incorrecta: %s", inicio)
     }

     if fin != grupo.FechaFin {
        t.Errorf("Fecha de fin incorrecta: %s", fin)
     }
}

// Comprueba que se regresa un error al editar un grupo inexistente
func TestEditaGrupoNoExiste(t *testing.T) {
     db := bddPrueba(t)
   
     defer db.Close()

     grupo := Grupo{
           ID:          1,
           Nombre:      "Metallica",
           FechaInicio: "28/10/1981",
           FechaFin:    "",
     } 

     err := EditaGrupo(grupo, db)

     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba un error, se obtuvo: %v", err)
     }
}

/*
Test para verificar que se agrega correctamente
una persona a un grupo
*/
func TestAgregaIntegrante(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)
     _, err := db.Exec(
        `INSERT INTO persons (
    		id_person,
    		stage_name,
    		real_name,
    		birth_date,
    		death_date
	) VALUES (1, 'Kurt Cobain', 'Kurt Donald Cobain',
	'20/02/1967', '05/04/1994');
        INSERT INTO groups VALUES (1, 'Metallica', '28/10/1981', '');
     `)
     if err != nil {
        t.Fatal(err)
     }

     err = AgregaIntegrante(1, 1, db)
     if err != nil {
        t.Fatalf("Error al agregar integrante: %v", err)
     }

     var idPersona, idGrupo int
     
     err = db.QueryRow(
         `SELECT id_person, id_group
         FROM in_group
     `).Scan(&idPersona, &idGrupo)

     if err != nil {
        t.Fatal(err)
     }

     if idPersona != 1 {
        t.Errorf("Id de persona incorrecto: %d", idPersona)
     }

     if idGrupo != 1 {
        t.Errorf("Id de grupo incorrecto: %d", idGrupo)
     }
}

/*
Prueba que verifica que se regrese un error
al intentar agregar una persona inexistente a un grupo
*/
func TestAgregaIntegrantePersonaNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)

     _, err := db.Exec(`
     	INSERT INTO groups VALUES (1, 'Metallica', '28/10/1981', '')
     `)
     if err != nil {
        t.Fatal(err)
     }

     err = AgregaIntegrante(1, 1, db)

     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba un error pero se obtuvo: %v", err)
     }
}

/*
Prueba que verifica que se lanze un error
al intentar agregar una persona a un grupo inexistente
*/
func TestAgregaIntegranteGrupoNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)

     _, err := db.Exec(`
        INSERT INTO persons (
    	       id_person,
    	       stage_name,
    	       real_name,
    	       birth_date,
    	       death_date
	) VALUES (1, 'Kurt Cobain', 'Kurt Donald Cobain',
	'20/02/1967', '05/04/1994')
     `)
     if err != nil {
        t.Fatal(err)
     }

     err = AgregaIntegrante(1, 1, db)

     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba un error pero se obtuvo: %v", err)
     }
}

// Prueba general para EliminaIntegrante
func TestEliminaIntegrante(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)

     _, err := db.Exec(`
     	INSERT INTO persons (
    		id_person,
    		stage_name,
    		real_name,
    		birth_date,
    		death_date
	) VALUES (1, 'Kurt Cobain', 'Kurt Donald Cobain',
	'20/02/1967', '05/04/1994');
        INSERT INTO groups VALUES (1, 'Metallica', '28/10/1981', '');

        INSERT INTO in_group VALUES (1, 1);
     `)
     if err != nil {
        t.Fatal(err)
     }

     err = EliminaIntegrante(1, 1, db)
     if err != nil {
        t.Fatalf("Error al eliminar integrante: %v", err)
     }

     var cantidad int

     err = db.QueryRow(`
         SELECT COUNT(*) FROM in_group
     `).Scan(&cantidad)
     if err != nil {
        t.Fatal(err)
     }

     if cantidad != 0 {
        t.Errorf("Se esperaba 0 integrantes pero se obtuvo: %d", cantidad)
     }
     err = db.QueryRow(`
         SELECT COUNT(*) FROM persons WHERE id_person = 1
     `).Scan(&cantidad)
     if err != nil {
        t.Fatal(err)
     }

     if cantidad != 1 {
        t.Errorf("La persona debería seguir existiendo, se obtuvo: %d", cantidad)
     }

     err = db.QueryRow(`
         SELECT COUNT(*) FROM groups WHERE id_group = 1
     `).Scan(&cantidad)
     if err != nil {
        t.Fatal(err)
     }

     if cantidad != 1 {
        t.Errorf("El grupo debería seguir existiendo, se obtuvo: %d", cantidad)
     }
}

/*
Prueba que verifica que se regrese un error
al intentar eliminar a una persona de un grupo
el cual no pertenece ni existe
*/
func TestEliminaIntegranteNoExiste(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)

     err := EliminaIntegrante(1, 1, db)

     if err != sql.ErrNoRows {
        t.Errorf("Se esperaba error pero se obtuvo: %v", err)
     }
}

// Test general para buscaIntegrantes
func TestBuscaIntegrantes(t *testing.T) {
     db := bddPrueba(t)
  
     defer db.Close()

     db.SetMaxOpenConns(1)

     _, err := db.Exec(
        `INSERT INTO persons VALUES
                (1, 'Fabricio', 'Federic Montes',
                '13/04/1978', '03/07/2008'),
                (2, 'Brian S.', 'Brian Soto Cruz',
                '14/11/1990', '');
        INSERT INTO in_group VALUES
               (1, 1),
               (2, 1);
     `)
     if err != nil {
        t.Fatal(err)
     }

     integrantes, err := BuscaIntegrantes(1, db)
     if err != nil {
        t.Fatalf("Error al buscar integrantes: %v", err)
     }

     if len(integrantes) != 2 {
        t.Fatalf("Se esperaban 2 integrantes, se obtuvieron: %d", len(integrantes))
     }

     if integrantes[0].ID != 1 {
        t.Errorf("ID incorrecto: %d", integrantes[0].ID)
     }

     if integrantes[0].NombreArtístico != "Fabricio" {
        t.Errorf("Nombre artístico incorrecto: %s", integrantes[0].NombreArtístico)
     }

     if integrantes[0].NombreReal != "Federic Montes" {
        t.Errorf("Nombre real incorrecto: %s", integrantes[0].NombreReal)
     }

    if integrantes[0].FechaNacimiento != "13/04/1978" {
        t.Errorf("Fecha de nacimiento incorrecta: %s", integrantes[0].FechaNacimiento)
    }

    if integrantes[0].FechaFallecimiento != "03/07/2008" {
        t.Errorf("Fecha de fallecimiento incorrecta: %s", integrantes[0].FechaFallecimiento)
    }

    if integrantes[1].NombreArtístico != "Brian S." {
        t.Errorf("Nombre artístico incorrecto: %s", integrantes[1].NombreArtístico)
    }
}

/*
Prueba para verificar el caso cuando el grupo no tiene integrantes
*/
func TestBuscaIntegrantesVacio(t *testing.T) {
     db := bddPrueba(t)
    
     defer db.Close()

     db.SetMaxOpenConns(1)

     integrantes, err := BuscaIntegrantes(1, db)
     if err != nil {
        t.Fatalf("No se esperaba un error: %v", err)
     }

     if len(integrantes) != 0 {
        t.Errorf("Se esperaba una lista vacía, se obtuvieron: %d integrantes", len(integrantes))
     }
}