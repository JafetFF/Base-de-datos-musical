package modelo

import (
       "database/sql"
       "fmt"
)

type Album struct {
     id      int
     ruta    string
     nombre  string
     año     int
}

type DatosAlbum struct {
     ID     int    `json:"id"`
     Ruta   string `json:"ruta"`
     Nombre string `json:"nombre"`
     Año    int	   `json:"año"`
}

func (a Album) Datos() DatosAlbum {
     return DatosAlbum{
     	    ID:     a.id,
	    Ruta:   a.ruta,
	    Nombre: a.nombre,
	    Año:    a.año,
     }
}

func (d DatosAlbum) ObtieneAlbumPorID(id int) Album {
     return Album{
            id:     id,
            ruta:   d.Ruta,
            nombre: d.Nombre,
            año:    d.Año,
     }
}

/*
Función que busca y regresa un album por su id
*/
func BuscaAlbumPorID(id int, db *sql.DB) (Album, error) {
     var album Album

     err := db.QueryRow(
         `SELECT id_album, path, name, year
          FROM albums
          WHERE id_album = ?`,
         id,
     ).Scan(
         &album.id,
         &album.ruta,
         &album.nombre,
         &album.año,
     )

     if err != nil {
        return Album{}, err
     }

     return album, nil
}

/*
Función que edita un álbum 
*/
func EditaAlbum(album Album, db *sql.DB) error {
     resultado, err := db.Exec(
         `UPDATE albums
          SET path = ?,
              name = ?,
              year = ?
          WHERE id_album = ?`,
         album.ruta,
         album.nombre,
         album.año,
         album.id,
     )

     if err != nil {
        return err
     }

     filas, err := resultado.RowsAffected()
     if err != nil {
        return err
     }

     if filas == 0 {
        return fmt.Errorf("No existe un álbum con el ID %d", album.id)
     }

     return nil
}

/*
Función que busca y regresa todos los álbumes de la BDD
*/
func BuscaAlbumes(db *sql.DB) ([]Album, error) {
     filas, err := db.Query(
            `SELECT id_album, path, name, year
            FROM albums
            ORDER BY name`)

     if err != nil {
        return nil, err
     }
     defer filas.Close()

     albumes := make([]Album, 0)

     for filas.Next() {
         var album Album

         err := filas.Scan(
             &album.id,
             &album.ruta,
             &album.nombre,
             &album.año,
         )

         if err != nil {
            return nil, err
         }

         albumes = append(albumes, album)
     }

     if err := filas.Err(); err != nil {
        return nil, err
     }

     return albumes, nil
}