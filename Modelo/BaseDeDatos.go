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
*/
func Agrega(rola Rola, db *sql.DB) error {
     _, err := db.Exec(
     	`INSERT INTO rolas (title, year, genre)
	VALUES (?, ?, ?)`, rola.title, rola.year, rola.genre)

     return err
}

/*
Función que elimina una rola de la bdd mediante su id
ya que el título podría repetirse
*/
func Elimina(id int, db *sql.DB) error {
     _, err := db.Exec(
     	`DELETE FROM rolas WHERE id_rola = ?`, id)

     return err
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
