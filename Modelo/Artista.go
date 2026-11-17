package modelo

import(
	"errors"
	"database/sql"
)

type Artista struct {
     Id	      int
     Nombre   string
     Tipo     int
}

// Función para corregir el artista de tipo desconocido a tipo artista o grupo
func AsignaPerformer(idRola, idPerformer int, db *sql.DB) error {
     var tipo int
     err := db.QueryRow(
         `SELECT id_type FROM performers WHERE id_performer = ?`,
         idPerformer,
     ).Scan(&tipo)

     if err != nil {
        return err
     }

     if tipo == 2 {
        return errors.New("el artista todavía es desconocido")
     }

     resultado, err := db.Exec(
     		`UPDATE rolas SET id_performer = ? WHERE id_rola = ?`,
        	idPerformer, idRola,
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