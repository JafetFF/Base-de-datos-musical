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

/*
Prueba Unitaria para el método del minero que lee etiquetas
*/
func TestLeeEtiquetas(t *testing.T) {     
     ruta := "./demo.mp3"

     datos, err := LeeEtiquetas(ruta)

     if err != nil {
         t.Fatal(err)
     }
     if datos.titulo != "T" {
     	t.Errorf("Título incorrecto: se esperaba T, se obtuvo %s",
	datos.titulo)
     }
     if datos.artista != "A" {
     	t.Errorf("Artista incorrecto: se esperaba A, se obtuvo %s",
	datos.artista)
     }
     if datos.album != "L" {
     	t.Errorf("Álbum incorrecto: se esperaba L, se obtuvo %s",
	datos.album)
     }
     if datos.año != 1950 {
     	t.Errorf("Año incorrecto: se esperaba 1950, se obtuvo %d",
	datos.año)
     }
     if datos.genero != "8" {
     	t.Errorf("Género incorrecto: se esperaba 8, se obtuvo %s",
	datos.genero)
     }
}


