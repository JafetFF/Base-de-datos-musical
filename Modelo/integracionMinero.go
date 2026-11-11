package main

import (
       "path/filepath"
       "Song/Minero"
       "database/sql"
       "fmt"
)

/*
Función que convierte los metadatos obtenidos
por el minero a una Rola
*/
func ConvierteMetadatos(datos minero.Metadatos) Rola {
    return Rola{
           artist:         datos.Artista,
           album:          datos.Album,
           albumPath:      filepath.Dir(datos.Ruta),
           title:          datos.Titulo,
           path:           datos.Ruta,
           track:          datos.Pista,
           totalTracks:    datos.TotalPistas,
           year:           datos.Año,
           genre:          datos.Genero,
           comments:       datos.Comentarios,
           composer:       datos.Compositor,
           originalArtist: datos.ArtistaOriginal,
           albumArtist:    datos.BandaAlbum,
           disc:           datos.NumeroDisco,
           totalDiscs:     datos.TotalDiscos,
           images:         datos.Imagenes,
    }
}

/*
Función que mina un directorio y agrega
las canciones encontradas a la base de datos
*/
func MinaYAgrega(directorio string, db *sql.DB) error {
     metadatos, err := minero.Mina(directorio)
     if err != nil {
         return err
     }

     for _, datos := range metadatos {
         rola := ConvierteMetadatos(datos)

         err := Agrega(rola, db)
         if err != nil {
             return fmt.Errorf("no se pudo agregar %q: %w",
	     rola.title, err)
         }
     }

     return nil
}