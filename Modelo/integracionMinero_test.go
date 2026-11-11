package main

import (
       "testing"
       "reflect"
       minero "Song/Minero"
)

/*
Prueba para comprobar que la conversión de tipos
se hace correctamente
*/ 
func TestConvierteMetadatos(t *testing.T) {
     imagenes := [][]byte{
         {1, 2, 3},
         {4, 5, 6},
     }

     datos := minero.NewMetadatos(minero.Metadatos{
         Titulo:          "Cancion",
         Artista:         "Artista",
         Album:           "Album",
         BandaAlbum:      "Banda",
         Año:            2024,
         Pista:           3,
         TotalPistas:     12,
         NumeroDisco:     1,
         TotalDiscos:     2,
         Genero:          "Rock",
         Comentarios:     "Comentario",
         Compositor:      "Compositor",
         ArtistaOriginal: "Artista original",
         Imagenes:        imagenes,
         Ruta:            "/musica/album/rola.mp3",
     })

     esperada := Rola{
         artist:         "Artista",
         album:          "Album",
         albumPath:      "/musica/album",
         title:          "Cancion",
         path:           "/musica/album/rola.mp3",
         track:          3,
         totalTracks:    12,
         year:           2024,
         genre:          "Rock",
         comments:       "Comentario",
         composer:       "Compositor",
         originalArtist: "Artista original",
         albumArtist:    "Banda",
         disc:           1,
         totalDiscs:     2,
         images:         imagenes,
     }

     obtenida := ConvierteMetadatos(datos)

     if !reflect.DeepEqual(obtenida, esperada) {
         t.Errorf("ConvierteMetadatos() = %+v, se esperaba %+v",
         obtenida, esperada)
     }
}

// Prueba para el método MinaYAgrega()
func TestMinaYAgrega(t *testing.T) {
     db := bddPrueba(t)

     err := MinaYAgrega("../Minero", db)
     if err != nil {
         t.Fatal(err)
     }

     rolas, err := Busca("T", db)
     if err != nil {
         t.Fatal(err)
     }

     if len(rolas) != 1 {
         t.Fatalf("Se esperaba 1 rola, se obtuvieron %d", len(rolas))
     }

     rola := rolas[0]

     if rola.title != "T" {
         t.Errorf("Título incorrecto: se esperaba T, se obtuvo %s", rola.title)
     }

     if rola.artist != "A" {
         t.Errorf("Artista incorrecto: se esperaba A, se obtuvo %s", rola.artist)
     }

     if rola.album != "L" {
         t.Errorf("Álbum incorrecto: se esperaba L, se obtuvo %s", rola.album)
     }

     if rola.year != 1950 {
         t.Errorf("Año incorrecto: se esperaba 1950, se obtuvo %d", rola.year)
     }
}