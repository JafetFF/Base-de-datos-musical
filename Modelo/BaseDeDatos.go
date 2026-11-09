package main

import (
       "database/sql"
       "fmt"
       "bufio"
       "strings"
       "os"
       _ "modernc.org/sqlite"
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

/*
Función principal que permite abrir la base de datos
*/
func main() {
     db, err := sql.Open("sqlite", "../BDD/BaseMusical.db")

     if err != nil {
     	fmt.Printf("Error %v al abrir la base de datos \n", err)
	return
     }
     defer db.Close()

     err = db.Ping()

     if err != nil {
     	fmt.Printf("Error %v al conectar a la base de datos", err)
	return
     }

     RecibeInstruccion(db)     
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
	 cadena := strings.Join(result[1:]," ")
	 switch result[0] {
	 case "search":
	    rolas, err := Busca(cadena, db)
	    if err != nil {
	       fmt.Printf("Error al buscar las rolas: %v\n", err)
	       return
	    }
	    MuestraRolas(rolas)
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
        `SELECT id_rola, title, year, genre
         FROM rolas
         WHERE title LIKE ?`,
        "%"+cadena+"%",
    )
    
    if err != nil {
        return nil, err
    }
    defer renglones.Close()

    var rolas []Rola

    for renglones.Next() {
        var rola Rola

        err := renglones.Scan(
            &rola.id,
            &rola.title,
            &rola.year,
            &rola.genre,
        )
        if err != nil {
            return nil, err
        }

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
     return tx.Commit()
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
     _, err := db.Exec(
     	`UPDATE rolas
	 SET title = ?,
 	     track = ?,
	     year  = ?,
	     genre = ?
	     WHERE id_rola = ?`,
	     rola.title, rola.track,
	     rola.year, rola.genre, id)
     return err
}

/*
Función que muestra los resultados encontrados en la bdd
*/
func MuestraRolas(rolas []Rola) {
     // Muestra rolas
     //for _, rola := range rolas {
     // cada cancion del arreglo recibido
     // fmt.Printf("%d | %s | %s | %", rola.id)
     //	    }
}
