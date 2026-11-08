package minero

import (
       "path/filepath"
       "strings"
       "io/fs"
       "os"
       "github.com/dhowden/tag"
)

/*
Estructura de los metadatos que se extraerán de cada rola
*/
type Metadatos struct {
     titulo          string
     artista         string
     album           string
     bandaAlbum      string
     año             int
     pista           int
     totalPistas     int
     numeroDisco     int
     totalDiscos     int
     genero          string
     comentarios     string
     compositor      string
     artistaOriginal string
     imagenes        [][]byte
     ruta            string
}

/*
Función que mina un directorio devolviendo los metadatos
para después almacenarlos en la base de datos
*/
func Mina(directorio string) ([]Metadatos, error) {
    rutas, err := BuscaMp3(directorio)
    if err != nil {
        return nil, err
    }

    resultados := make([]Metadatos, 0, len(rutas))

    for _, ruta := range rutas {
        datos, err := LeeEtiquetas(ruta)
        if err != nil {
            return nil, err
        }
        resultados = append(resultados, datos)
    }

    return resultados, nil
}

/*
Función que recibe un directorio y devuelve
los directorios de todos los archivos encontrados
*/
func BuscaMp3(directorio string) ([]string, error) {
   var canciones []string
   
   err := filepath.WalkDir(directorio,
   func(ruta string, entrada fs.DirEntry, err error) error {
   if err != nil {
      return err
   }

   if !entrada.IsDir() && strings.EqualFold(filepath.Ext(ruta), ".mp3") {
      canciones = append(canciones, ruta)
   }

   return nil
   
   })

   return canciones, err
}

/*
Función que lee las etiquetas de una rola pasándole su directorio
*/
func LeeEtiquetas(ruta string) (Metadatos, error) {
    var datos Metadatos

    archivo, err := os.Open(ruta)
    if err != nil {
       return Metadatos{}, err
    }
    defer archivo.Close()

    metadata, err := tag.ReadFrom(archivo)
    if err != nil {
       return Metadatos{}, err
    }

    datos.titulo = metadata.Title()
    datos.artista = metadata.Artist()
    datos.album = metadata.Album()
    datos.genero = metadata.Genre()
    datos.ruta = ruta
    datos.año = metadata.Year()
    datos.comentarios = metadata.Comment()
    datos.compositor = metadata.Composer()
    
    datos.pista, datos.totalPistas = metadata.Track()
    datos.numeroDisco, datos.totalDiscos = metadata.Disc()

    raw := metadata.Raw()

    if banda, ok := raw["TPE2"].(string); ok {
       datos.bandaAlbum = banda
    }
    if original, ok := raw["TOPE"].(string); ok {
       datos.artistaOriginal = original
    }

    for _, valor := range raw {
    	imagen, ok := valor.(*tag.Picture)
    	if !ok {
           continue
    	}
    	datos.imagenes = append(datos.imagenes, imagen.Data)
    }

    return datos, nil
}
