package minero

import (
       "testing"
       "path/filepath"
       "os"
)

/*
Prueba unitaria para el método que busca del minero
Aquí se crea un directorio temporal donde se crean
distintos archivos, y que posteriormente se regresen
los archivos MP3
*/
func TestBuscaMP3(t *testing.T) {
     directorio := t.TempDir()

     os.Mkdir(filepath.Join(directorio, "Rock"), 0755)
     
     os.Create(filepath.Join(directorio, "Rock", "rola1.mp3"))
     os.Create(filepath.Join(directorio, "Rock", "rola2.mp3"))
     os.Create(filepath.Join(directorio, "Rock", "imagen.png"))
     os.Create(filepath.Join(directorio, "rola3.mp3"))

     canciones, err := BuscaMp3(directorio)
     if err != nil {
	t.Fatal(err)
     }
     if len(canciones) != 3 {
	t.Fatalf("Se esperaban 3 canciones, se encontraron %d", len(canciones))
     }
}