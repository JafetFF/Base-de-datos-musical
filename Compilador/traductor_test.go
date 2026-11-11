package compilador

import (
       "testing"
       "slices"
)

type NodoTest struct {
}

func (n *NodoTest) esNodo() {}

// Prueba unitaria para el método VerificaNodo
func TestVerificaNodo(t *testing.T) {
     pruebas := []struct {
         nombre string
         campo string
         operador string
         valor string
	 tipoValor TipoToken
         errorEsperado bool
     }{
         {"Título igual", "title", "=", "Hero", CADENA, false},
         {"Título diferente", "title", "!=", "Liar", CADENA, false},
         {"Título contiene", "title", "~", "Arsonists", CADENA, false},
         {"Año mayor o igual", "year", ">=", "2004", NUMERO, false},
         {"Pista menor que", "track", "<", "5", NUMERO, false},
         {"Campo no permitido", "Unknown", "=", "Test", CADENA, true},
         {"Operador no permitido", "title", "?", "Test", CADENA, true},
         {"Operador incompatible", "title", ">", "Test", CADENA, true},
         {"Operador numérico en string", "genre", "<=", "Rock", CADENA, true},
         {"Contiene en año", "year", "~", "2017", NUMERO, true},
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             condicion := &Condicion{
                 Campo: prueba.campo,
                 Operador: prueba.operador,
                 Valor: prueba.valor,
		 TipoValor: prueba.tipoValor,
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
	 tipoValor TipoToken
         sqlEsperado string
     }{
         {"Título igual", "title", "=", "Dive", CADENA, "rolas.title = ?"},
         {"Año mayor o igual", "year", ">=", "2008", NUMERO, "rolas.year >= ?"},
         {"Género diferente", "genre", "!=", "Pop", CADENA, "rolas.genre != ?"},
         {"Artista igual", "artist", "=", "Eiro", CADENA, "performers.name = ?"},
         {"Álbum igual", "album", "=", "Reasons", CADENA, "albums.name = ?"},
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             condicion := &Condicion{
                 Campo: prueba.campo,
                 Operador: prueba.operador,
                 Valor: prueba.valor,
		 TipoValor: prueba.tipoValor,
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
	     TipoValor: CADENA,
         },
         Derecho: &Condicion{
             Campo: "year",
             Operador: ">=",
             Valor: "2017",
	     TipoValor: NUMERO,
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
	     TipoValor: CADENA,
         },
         Derecho: &Condicion{
             Campo: "genre",
             Operador: "=",
             Valor: "Pop",
	     TipoValor: CADENA,
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
		 TipoValor: CADENA,
             },
             Derecho: &Condicion{
                 Campo: "year",
                 Operador: ">=",
                 Valor: "2015",
		 TipoValor: NUMERO,
             },
         },
         Derecho: &Condicion{
             Campo: "genre",
             Operador: "=",
             Valor: "Rock",
	     TipoValor: CADENA,
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
             Campo:     "genre",
             Operador:  "=",
             Valor:     "Pop",
	     TipoValor: CADENA,
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
                 Campo:     "title",
                 Operador:  "=",
                 Valor:     "Guerrero",
		 TipoValor: CADENA,
             },
             Derecho: &Condicion{
                 Campo:     "year",
                 Operador:  ">=",
                 Valor:     "2009",
		 TipoValor: NUMERO,
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
                 Campo:     "genre",
                 Operador:  "=",
                 Valor:     "Indie",
		 TipoValor: CADENA,
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

/*
Prueba unitaria para verificar que efectivamente
use regresa un error el método Traduce() al
pasarle un Nodo desconocido
*/
func TestTraduceNodoDesconocido(t *testing.T) {
     traductor := Traductor{}
     nodo := &NodoTest{}

     _, _, err := traductor.Traduce(nodo)

     if err == nil {
         t.Error("Se esperaba un error por tipo de Nodo desconocido")
     }
}

/*
Prueba unitaria para comprobar que se traduce
bien el caso de LIKE en Traduce()
*/
func TestTraduceContiene(t *testing.T) {
     traductor := Traductor{}
     nodo := &Condicion{
     	  Campo:     "title",
          Operador:  "~",
          Valor:     "Heaven",
	  TipoValor: CADENA,
     }

     sql, parametros, err := traductor.Traduce(nodo)

     if err != nil {
         t.Fatalf("Error al traducir: %v", err)
     }

     esperado := "rolas.title LIKE ?"
     if sql != esperado {
         t.Errorf("SQL esperado %q, pero se obtuvo %q", esperado, sql)
     }

     parametrosEsperados := []any{"%Heaven%"}
     if !slices.Equal(parametros, parametrosEsperados) {
         t.Errorf("Parámetros esperados %v, y se obtuvieron %v",
         parametrosEsperados, parametros)
     }
}

/*
Prueba unitaria para comprobar que se regresa
un error cuando el tipo del valor no corresponde
con el que puede recibir el campo
*/
func TestVerificaTipo(t *testing.T) {
     pruebas := []struct {
         nombre string
         campo string
         tipo TipoToken
         esperado bool
     }{
         {"Cadena válida", "title", CADENA, false},
         {"Número válido", "year", NUMERO, false},
         {"Cadena en campo numérico", "year", CADENA, true},
         {"Número en campo de texto", "title", NUMERO, true},
         {"Campo desconocido", "unknown", CADENA, true},
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             condicion := &Condicion{
                 Campo:     prueba.campo,
                 TipoValor: prueba.tipo,
             }

             err := verificaTipo(condicion)

             if prueba.esperado && err == nil {
                 t.Error("Se esperaba un error")
             }

             if !prueba.esperado && err != nil {
                 t.Errorf("No se esperaba un error, pero se obtuvo: %v",
		 err)
             }
         })
     }
}

/*
Prueba unitaria para el método ConstruyeConsulta()
*/
func TestConstruyeConsulta(t *testing.T) {
     traductor := Traductor{}

     pruebas := []struct {
         nombre       string
         nodo         Nodo
         esperadoSQL  string
         esperadoArgs []any
     }{
         {
             nombre: "Sin JOIN",
             nodo: &Condicion{
                 Campo:     "title",
                 Operador:  "=",
                 Valor:     "Eyes On You",
                 TipoValor: CADENA,
             },
             esperadoSQL: `SELECT
    	      	  rolas.id_rola,
    	      	  COALESCE(rolas.id_performer, 0),
    	      	  COALESCE(rolas.id_album, 0),
    	      	  COALESCE(performers.name, ''),
    	      	  COALESCE(albums.name, ''),
    	      	  COALESCE(albums.path, ''),
    	      	  COALESCE(rolas.title, ''),
		  COALESCE(rolas.path, ''),
    	      	  COALESCE(rolas.track, 0),
    	      	  COALESCE(rolas.year, 0),
    	     	  COALESCE(rolas.genre, '')
	      	  FROM rolas
	      	  JOIN performers ON rolas.id_performer = performers.id_performer
	      JOIN albums ON rolas.id_album = albums.id_album WHERE rolas.title = ?`,
             esperadoArgs: []any{"Eyes On You"},
         },
         {
             nombre: "JOIN performers",
             nodo: &Condicion{
                 Campo:     "artist",
                 Operador:  "=",
                 Valor:     "Lovkn",
                 TipoValor: CADENA,
             },
             esperadoSQL: `SELECT
    	      	  rolas.id_rola,
    	      	  COALESCE(rolas.id_performer, 0),
    	      	  COALESCE(rolas.id_album, 0),
    	      	  COALESCE(performers.name, ''),
    	      	  COALESCE(albums.name, ''),
    	      	  COALESCE(albums.path, ''),
    	      	  COALESCE(rolas.title, ''),
		  COALESCE(rolas.path, ''),
    	      	  COALESCE(rolas.track, 0),
    	      	  COALESCE(rolas.year, 0),
    	     	  COALESCE(rolas.genre, '')
	      	  FROM rolas
	      	  JOIN performers ON rolas.id_performer = performers.id_performer
	      JOIN albums ON rolas.id_album = albums.id_album WHERE performers.name = ?`,
             esperadoArgs: []any{"Lovkn"},
         },
         {
             nombre: "JOIN albums",
             nodo: &Condicion{
                 Campo:     "album",
                 Operador:  "=",
                 Valor:     "Hits",
                 TipoValor: CADENA,
             },
             esperadoSQL: `SELECT
    	      	  rolas.id_rola,
    	      	  COALESCE(rolas.id_performer, 0),
    	      	  COALESCE(rolas.id_album, 0),
    	      	  COALESCE(performers.name, ''),
    	      	  COALESCE(albums.name, ''),
    	      	  COALESCE(albums.path, ''),
    	      	  COALESCE(rolas.title, ''),
		  COALESCE(rolas.path, ''),
    	      	  COALESCE(rolas.track, 0),
    	      	  COALESCE(rolas.year, 0),
    	     	  COALESCE(rolas.genre, '')
	      	  FROM rolas
	      	  JOIN performers ON rolas.id_performer = performers.id_performer
	      JOIN albums ON rolas.id_album = albums.id_album WHERE albums.name = ?`,
             esperadoArgs: []any{"Hits"},
         },
         {
             nombre: "Dos JOIN",
             nodo: &OperacionLogica{
                 Operador: "&",
                 Izquierdo: &Condicion{
                     Campo:     "artist",
                     Operador:  "=",
                     Valor:     "Alfonso Peduto",
                     TipoValor: CADENA,
                 },
                 Derecho: &Condicion{
                     Campo:     "album",
                     Operador:  "=",
                     Valor:     "Binary Data IX",
                     TipoValor: CADENA,
                 },
             },
             esperadoSQL: `SELECT
    	      	  rolas.id_rola,
    	      	  COALESCE(rolas.id_performer, 0),
    	      	  COALESCE(rolas.id_album, 0),
    	      	  COALESCE(performers.name, ''),
    	      	  COALESCE(albums.name, ''),
    	      	  COALESCE(albums.path, ''),
    	      	  COALESCE(rolas.title, ''),
		  COALESCE(rolas.path, ''),
    	      	  COALESCE(rolas.track, 0),
    	      	  COALESCE(rolas.year, 0),
    	     	  COALESCE(rolas.genre, '')
	      	  FROM rolas
	      	  JOIN performers ON rolas.id_performer = performers.id_performer
	      JOIN albums ON rolas.id_album = albums.id_album WHERE (performers.name = ? AND albums.name = ?)`,
             esperadoArgs: []any{"Alfonso Peduto", "Binary Data IX"},
         },
     }

     for _, prueba := range pruebas {
         t.Run(prueba.nombre, func(t *testing.T) {
             sql, args, err := traductor.ConstruyeConsulta(prueba.nodo)

             if err != nil {
                 t.Fatalf("Error al construir la consulta: %v", err)
             }

             if sql != prueba.esperadoSQL {
                 t.Errorf("SQL esperado: \n%s \nSQL obtenido: \n%s",
                 prueba.esperadoSQL, sql)
             }

             if !slices.Equal(args, prueba.esperadoArgs) {
                 t.Errorf("Parámetros esperados %v, obtenidos %v",
                 prueba.esperadoArgs, args)
             }
         })
     }
}