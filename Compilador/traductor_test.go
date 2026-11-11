package compilador

import (
       "testing"
       "slices"
)

// Prueba unitaria para el método VerificaNodo
func TestVerificaNodo(t *testing.T) {
     pruebas := []struct {
         nombre string
         campo string
         operador string
         valor string
         errorEsperado bool
     }{
         {"Título igual", "title", "=", "Hero", false},
         {"Título diferente", "title", "!=", "Liar", false},
         {"Título contiene", "title", "~", "Arsonists", false},
         {"Año mayor o igual", "year", ">=", "2004", false},
         {"Pista menor que", "track", "<", "5", false},
         {"Campo no permitido", "Unknown", "=", "Test", true},
         {"Operador no permitido", "title", "?", "Test", true},
         {"Operador incompatible", "title", ">", "Test", true},
         {"Operador numérico en string", "genre", "<=", "Rock", true},
         {"Contiene en año", "year", "~", "2017", true},
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             condicion := &Condicion{
                 Campo: prueba.campo,
                 Operador: prueba.operador,
                 Valor: prueba.valor,
             }

             _, err := verificaNodo(condicion)
             if prueba.errorEsperado && err == nil {
             	t.Errorf("Se esperaba un error")
             }

             if !prueba.errorEsperado && err != nil {
                t.Errorf("No se esperaba un error: %v", err)
             }
         })
     }
}

// Prueba unitaria que comprueba que se traduce bien una Condicion
func TestTraduceCondicion(t *testing.T) {
     traductor := Traductor{}

     pruebas := []struct {
         nombre string
         campo string
         operador string
         valor string
         sqlEsperado string
     }{
         {"Título igual", "title", "=", "Dive", "rolas.title = ?"},
         {"Año mayor o igual", "year", ">=", "2008", "rolas.year >= ?"},
         {"Género diferente", "genre", "!=", "Pop", "rolas.genre != ?"},
         {"Artista igual", "artist", "=", "Eiro", "performers.name = ?"},
         {"Álbum igual", "album", "=", "Reasons", "albums.name = ?"},
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             condicion := &Condicion{
                 Campo: prueba.campo,
                 Operador: prueba.operador,
                 Valor: prueba.valor,
             }

             sql, parametros, err := traductor.Traduce(condicion)

             if err != nil {
                 t.Fatalf("Error al traducir: %v", err)
             }

             if sql != prueba.sqlEsperado {
                 t.Errorf("SQL esperado %q, pero se obtuvo %q",
                 prueba.sqlEsperado, sql)
             }

             if len(parametros) != 1 {
                 t.Fatalf("Se esperaba 1 parámetro, y se obtuvieron %d",
		 len(parametros))
             }

             if parametros[0] != prueba.valor {
                 t.Errorf("Parámetro esperado %q, se obtuvo %v",
                 prueba.valor, parametros[0])
             } 
         })
     }
}

/*
Prueba unitaria para probar que se traduce
bien el caso de OperacionLogica con &
*/
func TestTraduceOperacionAND(t *testing.T) {
     traductor := Traductor{}

     nodo := &OperacionLogica{
         Operador: "&",
         Izquierdo: &Condicion{
             Campo: "title",
             Operador: "=",
             Valor: "Otra Vez",
         },
         Derecho: &Condicion{
             Campo: "year",
             Operador: ">=",
             Valor: "2017",
         },
     }

    sql, parametros, err := traductor.Traduce(nodo)
    if err != nil {
        t.Fatalf("Error al traducir: %v", err)
    }

    esperado := "(rolas.title = ? AND rolas.year >= ?)"
    if sql != esperado {
        t.Errorf("SQL esperado %q, pero se obtuvo %q", esperado, sql)
    }

    parametrosEsperados := []any{"Otra Vez", "2017"}
    if !slices.Equal(parametros, parametrosEsperados) {
        t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
	parametrosEsperados, parametros)
    }
}

/*
Prueba unitaria para probar que se traduce
bien el caso de OperacionLogica con |
*/
func TestTraduceOperacionOR(t *testing.T) {
     traductor := Traductor{}

     nodo := &OperacionLogica{
         Operador: "|",
         Izquierdo: &Condicion{
             Campo: "title",
             Operador: "=",
             Valor: "Changed",
         },
         Derecho: &Condicion{
             Campo: "genre",
             Operador: "=",
             Valor: "Pop",
         }, 
     }

     sql, parametros, err := traductor.Traduce(nodo)
     if err != nil {
         t.Fatalf("Error al traducir: %v", err)
     }

     esperado := "(rolas.title = ? OR rolas.genre = ?)"
     if sql != esperado {
         t.Errorf("SQL esperado %q, y se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"Changed", "Pop"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
	 parametrosEsperados, parametros)
     }
}

/*
Prueba unitaria para comprobar que una
OperacionLogica con & y | se traducen bien
*/
func TestTraduceOperacionesAnidadas(t *testing.T) {
     traductor := Traductor{}

     nodo := &OperacionLogica{
         Operador: "|",
         Izquierdo: &OperacionLogica{
             Operador: "&",
             Izquierdo: &Condicion{
                 Campo: "title",
                 Operador: "=",
                 Valor: "Lord",
             },
             Derecho: &Condicion{
                 Campo: "year",
                 Operador: ">=",
                 Valor: "2015",
             },
         },
         Derecho: &Condicion{
             Campo: "genre",
             Operador: "=",
             Valor: "Rock",
         },
     }

     sql, parametros, err := traductor.Traduce(nodo)
     if err != nil {
         t.Fatalf("Error al traducir: %v", err)
     }

     esperado := "((rolas.title = ? AND rolas.year >= ?) OR rolas.genre = ?)"
     if sql != esperado {
         t.Errorf("SQL esperado %q, se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"Lord", "2015", "Rock"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, pero se obtuvieron %v",
	 parametrosEsperados, parametros)
     }
}

/*
Prueba unitaria para comprobar que se
traduce correctamente un caso simple
*/
func TestTraduceNegacion(t *testing.T) {
     traductor := Traductor{}

     nodo := &Negacion{
         Expresion: &Condicion{
             Campo:    "genre",
             Operador: "=",
             Valor:    "Pop",
         },
     }

     sql, parametros, err := traductor.Traduce(nodo)
     if err != nil {
         t.Fatalf("Error al traducir: %v", err)
     }

     esperado := "NOT (rolas.genre = ?)"
     if sql != esperado {
         t.Errorf("SQL esperado %q, pero se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"Pop"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
	 parametrosEsperados, parametros)
     }
}

/*
Prueba unitaria para verificar que se traduce
correctamente el caso de negación y AND
*/
func TestTraduceNegacionAND(t *testing.T) {
     traductor := Traductor{}

     nodo := &Negacion{
         Expresion: &OperacionLogica{
             Operador: "&",
             Izquierdo: &Condicion{
                 Campo:    "title",
                 Operador: "=",
                 Valor:    "Guerrero",
             },
             Derecho: &Condicion{
                 Campo:    "year",
                 Operador: ">=",
                 Valor:    "2009",
             },
         },
     }

     sql, parametros, err := traductor.Traduce(nodo)
     if err != nil {
         t.Fatalf("Error inesperado: %v", err)
     }

     esperado := "NOT ((rolas.title = ? AND rolas.year >= ?))"
     if sql != esperado {
         t.Errorf("SQL esperado %q, pero se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"Guerrero", "2009"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
         parametrosEsperados, parametros)
     }
}

/*
Prueba unitaria para verificar que se traduce
correctamente el caso de tener más de una Negacion
*/
func TestTraduceNegacionesMultiple(t *testing.T) {
     traductor := Traductor{}

     nodo := &Negacion{
         Expresion: &Negacion{
             Expresion: &Condicion{
                 Campo:    "genre",
                 Operador: "=",
                 Valor:    "Indie",
             },
         },
     }

     sql, parametros, err := traductor.Traduce(nodo)
     if err != nil {
         t.Fatalf("Error al traducir: %v", err)
     }

     esperado := "NOT (NOT (rolas.genre = ?))"
     if sql != esperado {
         t.Errorf("SQL esperado %q, pero se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"Indie"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
         parametrosEsperados, parametros)
     }
}