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
     if datos.Titulo != "T" {
     	t.Errorf("Título incorrecto: se esperaba T, se obtuvo %s",
	datos.Titulo)
     }
     if datos.Artista != "A" {
     	t.Errorf("Artista incorrecto: se esperaba A, se obtuvo %s",
	datos.Artista)
     }
     if datos.Album != "L" {
     	t.Errorf("Álbum incorrecto: se esperaba L, se obtuvo %s",
	datos.Album)
     }
     if datos.Año != 1950 {
     	t.Errorf("Año incorrecto: se esperaba 1950, se obtuvo %d",
	datos.Año)
     }
     if datos.Comentarios != "C" {
     	t.Errorf("Comentarios incorrectos: se esperaba C, se obtuvo %s",
	datos.Comentarios)
     }
     if datos.Genero != "8" {
     	t.Errorf("Género incorrecto: se esperaba 8, se obtuvo %s",
	datos.Genero)
     }
     if datos.Compositor != "S" {
     	t.Errorf("Compositor incorrecto: se esperaba S, se obtuvo %s",
	datos.Compositor)
     }
     if datos.Ruta != ruta {
     	t.Errorf("Ruta incorrecto: se esperaba %s, se obtuvo %s", ruta,
	datos.Ruta)
     }
     if datos.Pista != 3 {
        t.Errorf("Pista incorrecta: se esperaba 3, se obtuvo %d",
	datos.Pista)
     }
     if datos.TotalPistas != 7 {
     	t.Errorf("Total de pistas incorrecto: se esperaba 7, se obtuvo %d",
	datos.TotalPistas)
     }
     if datos.NumeroDisco != 1 {
     	t.Errorf("Número de disco incorrecto: se esperaba 1, se obtuvo %d",
	datos.NumeroDisco)
     }
     if datos.TotalDiscos != 0 {
     	t.Errorf("Total de discos incorrecto: se esperaba 0, se obtuvo %d",
	datos.TotalDiscos)
     }
     if datos.BandaAlbum != "B" {
     	t.Errorf("Banda del álbum incorrecta: se esperaba B, se obtuvo %s",
	datos.BandaAlbum)
     }
     if datos.ArtistaOriginal != "O" {
     	t.Errorf("Artista original incorrecto: se esperaba O, se obtuvo %s",
	datos.ArtistaOriginal)
     }
     if len(datos.Imagenes) != 2 {
     	t.Errorf("Cantidad de imágenes incorrecta: se esperaban 2, " +
	"se obtuvieron %d", len(datos.Imagenes))
     }
     for i, imagen := range datos.Imagenes {
     	 if len(imagen) == 0 {
            t.Errorf("La imagen %d está vacía", i)
	 }
     }
}


