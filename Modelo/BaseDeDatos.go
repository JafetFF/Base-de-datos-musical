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
     	fmt.Printf("Error %v al abrir la base de datos\n", err)
	return
     }
     defer db.Close()

     err = db.Ping()

     if err != nil {
     	fmt.Printf("Error %v al conectar a la base de datos", err)
	return
     }

     fmt.Print("si funciona")

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
	    MuestraRolas(Busca(cadena, db))
	 case "add":
	    //AgregaRola(cadena)
	 case "remove":
	    //EliminaRola(cadena)
	 case "edit":
	    //EditaRola(cadena)
	 }
     }
}

/*
Función que busca las coincidencias con la cadena recibida
en la bdd y luego las regresa
*/
func Busca(cadena string, db *sql.DB) ([]Rola) {
     //
     return nil
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