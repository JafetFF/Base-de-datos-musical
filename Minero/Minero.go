package minero

import (
       "path/filepath"
       "strings"
       "io/fs"
       "os"
       "github.com/dhowden/tag"
       "fmt"
)

/*
Estructura de los metadatos que se extraerán de cada rola
*/
type Metadatos struct {
     Titulo          string
     Artista         string
     Album           string
     BandaAlbum      string
     Año             int
     Pista           int
     TotalPistas     int
     NumeroDisco     int
     TotalDiscos     int
     Genero          string
     Comentarios     string
     Compositor      string
     ArtistaOriginal string
     Imagenes        [][]byte
     Ruta            string
}

// Constructor de los Metadatos
func NewMetadatos(d Metadatos) Metadatos {
    return Metadatos{
        Titulo:          d.Titulo,
        Artista:         d.Artista,
        Album:           d.Album,
        BandaAlbum:      d.BandaAlbum,
        Año:             d.Año,
        Pista:           d.Pista,
        TotalPistas:     d.TotalPistas,
        NumeroDisco:     d.NumeroDisco,
        TotalDiscos:     d.TotalDiscos,
        Genero:          d.Genero,
        Comentarios:     d.Comentarios,
        Compositor:      d.Compositor,
        ArtistaOriginal: d.ArtistaOriginal,
        Imagenes:        d.Imagenes,
        Ruta:            d.Ruta,
    }
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

    datos.Titulo = metadata.Title()
    datos.Artista = metadata.Artist()
    datos.Album = metadata.Album()
    datos.Genero = metadata.Genre()
    datos.Ruta = ruta
    datos.Año = metadata.Year()
    datos.Comentarios = metadata.Comment()
    datos.Compositor = metadata.Composer()
    
    datos.Pista, datos.TotalPistas = metadata.Track()
    datos.NumeroDisco, datos.TotalDiscos = metadata.Disc()

    raw := metadata.Raw()

    if banda, ok := raw["TPE2"].(string); ok {
       datos.BandaAlbum = banda
    }
    if original, ok := raw["TOPE"].(string); ok {
       datos.ArtistaOriginal = original
    }

    for _, valor := range raw {
    	imagen, ok := valor.(*tag.Picture)
    	if !ok {
           continue
    	}
    	datos.Imagenes = append(datos.Imagenes, imagen.Data)
	fmt.Println(
    "Tipo:", imagen.Type,
    "Descripción:", imagen.Description,
    "Formato:", imagen.MIMEType,
    "Tamaño:", len(imagen.Data),
)
    }

    return datos, nil
}
