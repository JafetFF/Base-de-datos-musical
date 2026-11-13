package modelo

import (
       "database/sql"
       "fmt"
       "bufio"
       "strings"
       "os"
       _ "modernc.org/sqlite"
       "Song/Compilador"
       "path/filepath"
       "net/http"
)

/*
Interface que permite utilizar
*sql.DB y *sql.Tx en las funciones
necesarias 
*/
type EjecutaSQL interface {
     Exec(query string, args ...any) (sql.Result, error)
     QueryRow(query string, args ...any) *sql.Row
}

// Función que recibe una instrucción para ejecutar en la bdd
func RecibeInstruccion(db *sql.DB) {
     scanner := bufio.NewScanner(os.Stdin)
     for scanner.Scan() {
     	 result := strings.Fields(scanner.Text())
	 if len(result) < 2 {
	    fmt.Println("Debes indicar una acción a realizar con al menos una cadena")
	    return
	 }
	 // cadena := strings.Join(result[1:]," ")
	 switch result[0] {
	 case "search": /*
	    rolas, err := Busca(cadena, db)
	    if err != nil {
	       fmt.Printf("Error al buscar las rolas: %v\n", err)
	       return
	    }*/
	 case "add":
	 // de alguna manera recibe la rola
	    // Agrega(rola, db)
	 case "remove":
	    //Elimina(rola, db)
	 case "edit":
	    //Edita(rola, db)
	 }
     }
}

/*
Función que busca las coincidencias con la cadena recibida
en la bdd y luego las regresa
*/
func Busca(cadena string, db *sql.DB) ([]Rola, error) {
    renglones, err := db.Query(
        `SELECT rolas.id_rola, rolas.id_album, rolas.title, rolas.year, rolas.genre, 
	 performers.name, albums.name, rolas.path
         FROM rolas
	 LEFT JOIN performers ON rolas.id_performer = performers.id_performer
	 LEFT JOIN albums ON rolas.id_album = albums.id_album
         WHERE rolas.title LIKE ?`,
        "%"+cadena+"%",
    )
    
    if err != nil {
        return nil, err
    }
    defer renglones.Close()

    var rolas []Rola

    for renglones.Next() {
        var rola Rola

        var artista, album, ruta sql.NullString
	var idAlbum sql.NullInt64

	err := renglones.Scan(
    	    &rola.id,
	    &idAlbum,
    	    &rola.title,
    	    &rola.year,
    	    &rola.genre,
    	    &artista,
    	    &album,
	    &ruta,
	)

	if err != nil {
    	   return nil, err
	}

	if idAlbum.Valid {
	   rola.idAlbum = int(idAlbum.Int64)
	}

	rola.artist = artista.String
	rola.album = album.String
	rola.path = ruta.String

        rolas = append(rolas, rola)
    }

    if err := renglones.Err(); err != nil {
        return nil, err
    }

    return rolas, nil
}


/*
Función que agrega una rola a la bdd
asegurándose de que el artista no esté
ya agregado en la base de datos
*/
func Agrega(rola Rola, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
     	return err
     }

     defer tx.Rollback()

     if strings.TrimSpace(rola.path) != "" {
     	existe, err := ExisteRola(rola.path, tx)
     	if err != nil {
     	   return nil
        }
     	if existe {
     	   return nil
        }
     }
     idPerformer, err := ObtienePerformer(rola, tx)
     if err != nil {
     	return err
     }

     idAlbum, err := ObtieneAlbum(rola, tx)
     if err != nil {
     	return err
     }

     _, err = tx.Exec(
        `INSERT INTO rolas (id_performer, id_album, title, path, track, year, genre)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
        idPerformer,
	idAlbum,
        rola.title,
        rola.path,
        rola.track,
        rola.year,
        rola.genre,
     )

     if err != nil {
     	return err
     }
     
     err = tx.Commit()
     if err != nil {
     	return err
     }
     return GuardaPortada(idAlbum, rola.images)
}

/*
Función auxiliar que verifica si una rola ya se encuentra en la bdd
*/
func ExisteRola(ruta string, db EjecutaSQL) (bool, error) {
     var existe bool

     err := db.QueryRow(
         `SELECT EXISTS(SELECT 1 FROM rolas WHERE path = ?)`, ruta,
     ).Scan(&existe)

     return existe, err
}

/*
Función auxiliar que agrega al artista si
no existe y que regresa el id del artista
*/
func ObtienePerformer(rola Rola, db EjecutaSQL) (int, error) {
     if strings.TrimSpace(rola.artist) == "" {
     	rola.artist = "Unknown"
     }
     id, err := BuscaPerformer(rola.artist, 2, db)

     if err == sql.ErrNoRows {
         resultado, err := db.Exec(
             `INSERT INTO performers (name, id_type)
              VALUES (?, ?)`,
              rola.artist, 2,
         )
         if err != nil {
             return 0, err
         }
         nuevoID, err := resultado.LastInsertId()
         if err != nil {
             return 0, err
         }
         return int(nuevoID), nil
     }
     return id, err
}

/*
Función que regresa el id de un artista
*/
func BuscaPerformer(nombre string, tipo int, db EjecutaSQL) (int, error) {
     var id int
     err := db.QueryRow(
         `SELECT id_performer
         FROM performers
         WHERE name = ? AND id_type = ?`,
         nombre, tipo,
     ).Scan(&id)

     return id, err
}

/*
Función auxiliar que agrega el álbum si no
existe y que regresa el id del álbum
*/
func ObtieneAlbum(rola Rola, db EjecutaSQL) (int, error) {
     if rola.album == "" {
     	return 0, nil
     }

     id, err := BuscaAlbum(rola.album, rola.albumPath, db)

     if err == sql.ErrNoRows {
         resultado, err := db.Exec(
             `INSERT INTO albums (name, path, year)
              VALUES (?, ?, ?)`,
              rola.album, rola.albumPath, rola.year,
         )
         if err != nil {
             return 0, err
         }
         nuevoID, err := resultado.LastInsertId()
         if err != nil {
             return 0, err
         }
         return int(nuevoID), nil
     }
     return id, err
}

// Guarda la portada de un álbum
func GuardaPortada(idAlbum int, imagenes [][]byte) error {
     if idAlbum == 0 || len(imagenes) == 0 {
        return nil
     }

     directorio := filepath.Join("BDD", "Portadas")

     err := os.MkdirAll(directorio, 0755)
     if err != nil {
        return err
     }

     imagen := imagenes[0]
     tipo := http.DetectContentType(imagen)

     extension := ""
     switch tipo {
     case "image/jpeg":
          extension = ".jpg"
	  
     case "image/png":
          extension = ".png"
	  
     case "image/gif":
          extension = ".gif"
	  
     case "image/webp":
          extension = ".webp"
	  
     default:
        return fmt.Errorf("formato de portada incompatible: %s", tipo)
     }

     nombre := fmt.Sprintf("%d%s", idAlbum, extension)
     
     ruta := filepath.Join(directorio, nombre)

     if _, err := os.Stat(ruta); err == nil {
        return nil
     }

     return os.WriteFile(ruta, imagen, 0644)
}

/*
Función que regresa el id de un álbum
*/
func BuscaAlbum(nombre string, ruta string, db EjecutaSQL) (int, error) {
     var id int

     err := db.QueryRow(
         `SELECT id_album
          FROM albums
          WHERE name = ? AND path = ?`,
          nombre, ruta,
     ).Scan(&id)

     return id, err
}

/*
Función que elimina una rola de la bdd mediante su id
ya que el título podría repetirse
*/
func Elimina(id int, db *sql.DB) error {
     result, err := db.Exec(
     	`DELETE FROM rolas WHERE id_rola = ?`, id)

     if err != nil {
     	return err
     }

     filasModificadas, err := result.RowsAffected()
     if err != nil {
	return err
     }
     if filasModificadas == 0 {
     	return fmt.Errorf("No existe un rola con el ID %d", id)
     }    

     return nil
}

/*
Función que edita una rola de la bdd mediante su id
y una rola que se le pasa como parámetro
supone que la rola existe en la bdd
*/
func Edita(id int, rola Rola, db *sql.DB) error {
     resultado, err := db.Exec(
     	`UPDATE rolas
	 SET title = ?,
 	     track = ?,
	     year  = ?,
	     genre = ?
	     WHERE id_rola = ?`,
	     rola.title, rola.track,
	     rola.year, rola.genre, id)
     if err != nil {
     	return err
     }
     filasModificadas, err := resultado.RowsAffected()
     if err != nil {
     	return err
     }
     if filasModificadas == 0 {
     	return fmt.Errorf("No exite una rola con el ID %d", id)
     }
     return nil
}

/*
Función que usa el compilador para hacer una
búsqueda y ejecutarla en la base de datos
*/
func BuscaCompilada(cadena string, db *sql.DB) ([]Rola, error) {
     lexer := compilador.NewLexer(cadena)
     parser, err := compilador.NewParser(lexer)
     if err != nil {
         return nil, err
     }

     nodo, err := parser.Parsea()
     if err != nil {
         return nil, err
     }

     traductor := compilador.Traductor{}
     consulta, parametros, err := traductor.ConstruyeConsulta(nodo)
     if err != nil {
         return nil, err
     }

     filas, err := db.Query(consulta, parametros...)
     if err != nil {
         return nil, err
     }
     defer filas.Close()

     rolas := make([]Rola, 0) 

     for filas.Next() {
       	 var rola Rola

    	 err := filas.Scan(
             &rola.id,
             &rola.idPerformer,
             &rola.idAlbum,
             &rola.artist,
             &rola.album,
             &rola.albumPath,
             &rola.title,
             &rola.path,
	     &rola.track,
             &rola.year,
             &rola.genre,
     	 )

    	 if err != nil {
            return nil, err
    	 }

    	 rolas = append(rolas, rola)
     }

    if err := filas.Err(); err != nil {
       return nil, err
    }

    return rolas, nil
}

/*
Busca y regresa todos los artistas en la BDD
*/
func BuscaArtistas(db *sql.DB) ([]Artista, error) {
     filas, err := db.Query(
            `SELECT id_performer, name, id_type
            FROM performers ORDER BY name`,
     )

     if err != nil {
        return nil, err
     }
     
     defer filas.Close()

     var artistas []Artista

     for filas.Next() {
         var artista Artista

         err := filas.Scan(
             &artista.Id,
             &artista.Nombre,
             &artista.Tipo,
         )

         if err != nil {
            return nil, err
         }

         artistas = append(artistas, artista)
     }

     if err := filas.Err(); err != nil {
        return nil, err
     }

     return artistas, nil
}

/*
Agrega una persona a la BDD como persona y como artista
*/
func AgregaPersona(persona Persona, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
        return err
     }

     defer tx.Rollback()

     nombre := persona.NombreArtístico
     if strings.TrimSpace(nombre) == "" {
     	nombre = persona.NombreReal
     }

     resultado, err := tx.Exec(
      	   	`INSERT INTO performers (name, id_type)
        	VALUES (?, ?)`, nombre, 0)
     if err != nil {
        return err
     }

     id, err := resultado.LastInsertId()
     if err != nil {
        return err
     }

     _, err = tx.Exec(
     	`INSERT INTO persons (
          	id_person, stage_name, real_name,
            	birth_date, death_date
         )
         VALUES (?, ?, ?, ?, ?)`,
         id,
         persona.NombreArtístico,
         persona.NombreReal,
         persona.FechaNacimiento,
         persona.FechaFallecimiento,
     )
     if err != nil {
        return err
     }

     return tx.Commit()
}

/*
Busca una persona por su Id
*/
func BuscaPersona(id int, db *sql.DB) (Persona, error) {
     var persona Persona

     err := db.QueryRow(
         `SELECT id_person, stage_name, real_name,
         birth_date, death_date
         FROM persons
         WHERE id_person = ?`, id,
     ).Scan(
         &persona.ID,
         &persona.NombreArtístico,
         &persona.NombreReal,
         &persona.FechaNacimiento,
         &persona.FechaFallecimiento,
     )

     if err != nil {
        return Persona{}, err
     }
     return persona, nil
}

/*
Edita una persona en la BDD
*/
func EditaPersona(persona Persona, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
        return err
     }

     defer tx.Rollback()
     resultado, err := tx.Exec(
        	`UPDATE performers
         	SET name = ?
         	WHERE id_performer = ?`,
        	persona.NombreArtístico,
        	persona.ID,
     )
     if err != nil {
        return err
     }
     filas, err := resultado.RowsAffected()
     if err != nil {
        return err
     }
     if filas == 0 {
        return fmt.Errorf("no existe el intérprete con la Id %d", persona.ID)
     }

     resultado, err = tx.Exec(
        	`UPDATE persons
         	SET stage_name = ?,
             	real_name = ?,
             	birth_date = ?,
             	death_date = ?
         	WHERE id_person = ?`,
        	persona.NombreArtístico,
        	persona.NombreReal,
        	persona.FechaNacimiento,
        	persona.FechaFallecimiento,
        	persona.ID,
     )
     if err != nil {
        return err
     }
     filas, err = resultado.RowsAffected()
     if err != nil {
        return err
     }
     if filas == 0 {
        return fmt.Errorf("no existe la persona con la Id %d", persona.ID)
     }

     return tx.Commit()
}

/*
Agrega un grupo a la BDD, se registra
como grupo pero también como artista
*/
func AgregaGrupo(grupo Grupo, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
        return err
     }

     defer tx.Rollback()

     resultado, err := tx.Exec(
        	`INSERT INTO performers (name, id_type)
         	VALUES (?, ?)`,
        	grupo.Nombre, 1,
     )
     if err != nil {
        return err
     }

     id, err := resultado.LastInsertId()
     if err != nil {
        return err
     }

     _, err = tx.Exec(
        `INSERT INTO groups (id_group, name, start_date, end_date)
        VALUES (?, ?, ?, ?)`,
        id,
        grupo.Nombre,
        grupo.FechaInicio,
        grupo.FechaFin,
     )
     if err != nil {
        return err
     }

     return tx.Commit()
}

/*
Busca un grupo en la BDD
con ayuda de su Id
*/
func BuscaGrupo(id int, db *sql.DB) (Grupo, error) {
     var grupo Grupo

     err := db.QueryRow(
         `SELECT id_group, name, start_date, end_date
          FROM groups
          WHERE id_group = ?`, id,
     ).Scan(
        &grupo.ID,
        &grupo.Nombre,
        &grupo.FechaInicio,
        &grupo.FechaFin,
     )
     if err != nil {
        return Grupo{}, err
     }

     return grupo, nil
}

/*
Edita los datos de un grupo en la BDD
*/
func EditaGrupo(grupo Grupo, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
        return err
     }

     defer tx.Rollback()

     resultado, err := tx.Exec(
        	`UPDATE performers
         	SET name = ?
         	WHERE id_performer = ?`,
        	grupo.Nombre,
        	grupo.ID,
     )
     if err != nil {
        return err
     }
     filas, err := resultado.RowsAffected()
     if err != nil {
        return err
     }

     if filas == 0 {
        return sql.ErrNoRows
     }

     resultado, err = tx.Exec(
        	`UPDATE groups
         	SET name = ?, start_date = ?, end_date = ?
         	WHERE id_group = ?`,
        	grupo.Nombre,
        	grupo.FechaInicio,
        	grupo.FechaFin,
        	grupo.ID,
     )
     if err != nil {
        return err
     }
     filas, err = resultado.RowsAffected()
     if err != nil {
        return err
     }
     if filas == 0 {
        return sql.ErrNoRows
     }

     return tx.Commit()
}

/*
Agrega una persona como un integrante de un grupo
*/
func AgregaIntegrante(idPersona, idGrupo int, db *sql.DB) error {
     tx, err := db.Begin()
     if err != nil {
        return err
     }

     defer tx.Rollback()

     var existe bool

     err = tx.QueryRow(
         `SELECT EXISTS(
                 SELECT 1 FROM persons WHERE id_person = ?
         )`, idPersona,
     ).Scan(&existe)

     if err != nil {
        return err
     }

     if !existe {
        return sql.ErrNoRows
     }

     err = tx.QueryRow(
         `SELECT EXISTS(
                 SELECT 1 FROM groups WHERE id_group = ?
         )`, idGrupo,
     ).Scan(&existe)

     if err != nil {
        return err
     }

     if !existe {
        return sql.ErrNoRows
     }

     _, err = tx.Exec(
        `INSERT INTO in_group (id_person, id_group)
        VALUES (?, ?)`, idPersona, idGrupo)

     if err != nil {
        return err
     }

     return tx.Commit()
}

/*
Elimina una persona de un grupo
*/
func EliminaIntegrante(idPersona, idGrupo int, db *sql.DB) error {
     resultado, err := db.Exec(
                `DELETE FROM in_group
         	WHERE id_person = ? AND id_group = ?`,
        	idPersona,
        	idGrupo,
     )
     if err != nil {
        return err
     }
     filas, err := resultado.RowsAffected()
     if err != nil {
        return err
     }

     if filas == 0 {
        return sql.ErrNoRows
     }
     return nil
}

/*
Busca a los integrantes de un grupo con ayuda de su id
*/
func BuscaIntegrantes(idGrupo int, db *sql.DB) ([]Persona, error) {
     filas, err := db.Query(
            `SELECT persons.id_person,
                    persons.stage_name,
                    persons.real_name,
                    persons.birth_date,
                    persons.death_date
         	    FROM persons
         	    JOIN in_group
         	    ON persons.id_person = in_group.id_person
         	    WHERE in_group.id_group = ?
		    ORDER BY persons.id_person`, idGrupo)
     if err != nil {
        return nil, err
     }
     defer filas.Close()

     var integrantes []Persona

     for filas.Next() {
         var persona Persona

         err := filas.Scan(
             &persona.ID,
             &persona.NombreArtístico,
             &persona.NombreReal,
             &persona.FechaNacimiento,
             &persona.FechaFallecimiento,
         )
         if err != nil {
            return nil, err
         }

         integrantes = append(integrantes, persona)
     }

     if err := filas.Err(); err != nil {
        return nil, err
     }

     return integrantes, nil
}

/*
Busca y regresa a todas las personas registradas en la BDD
*/
func BuscaPersonas(db *sql.DB) ([]Persona, error) {
     filas, err := db.Query(`
            SELECT id_person,
                   stage_name,
                   real_name,
                   birth_date,
                   death_date
                   FROM persons
                   ORDER BY id_person`)
     if err != nil {
        return nil, err
     }
     defer filas.Close()

     personas := make([]Persona, 0)

     for filas.Next() {
         var persona Persona

         err := filas.Scan(
             &persona.ID,
             &persona.NombreArtístico,
             &persona.NombreReal,
             &persona.FechaNacimiento,
             &persona.FechaFallecimiento,
         )
         if err != nil {
            return nil, err
         }

         personas = append(personas, persona)
     }

     if err := filas.Err(); err != nil {
        return nil, err
     }

     return personas, nil
}