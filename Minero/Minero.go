package minero

import (
       "path/filepath"
       "strings"
       "io/fs"
)

/*
Función que recibe un directorio y devuelve
los directorios de todos los archivos encontrados
*/
func BuscaMp3(directorio string) ([]string, error){
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
