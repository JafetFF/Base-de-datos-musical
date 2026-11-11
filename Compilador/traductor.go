package compilador

import (
       "fmt"
       "slices"
)

type Traductor struct {
}

// Campos que se pueden recibir para traducir a SQL
var camposPermitidos = map[string]string{
    "title":       "rolas.title",
    "track":       "rolas.track",
    "year":        "rolas.year",
    "genre":       "rolas.genre",
    "path":        "rolas.path",
    "artist":      "performers.name",
    "album":       "albums.name",
    "album_path":  "albums.path",
    "album_year":  "albums.year",
}

// Operadores que se pueden recibir para realizar búsquedas
var operadoresPermitidos = map[string]string{
    "=":  "=",
    "!=": "!=",
    ">":  ">",
    "<":  "<",
    ">=": ">=",
    "<=": "<=",
    "~":  "LIKE",
}

// Correspondencia de operador por campo 
var operadoresPorCampo = map[string][]string{
    "title":  {"=", "!=", "~"},
    "artist": {"=", "!=", "~"},
    "album":  {"=", "!=", "~"},
    "genre":  {"=", "!=", "~"},
    "year":   {"=", "!=", ">", "<", ">=", "<="},
    "track":  {"=", "!=", ">", "<", ">=", "<="},
}

// Tipos para cada campo
var tiposPorCampo = map[string]TipoToken{
    "title":       CADENA,
    "artist":	   CADENA,
    "year":	   NUMERO,
    "track":       NUMERO,
    "genre":       CADENA,
    "path":        CADENA,
    "album":       CADENA,
    "album_path":  CADENA,
    "album_year":  NUMERO,
}

// Los posibles Joins que se pueden hacer
var joinsPosibles = map[string]string{
    "performers": "JOIN performers ON rolas.id_performer = performers.id_performer",
    "albums":     "JOIN albums ON rolas.id_album = albums.id_album",
}

// Correspondencia de joins por campo
var joinsPorCampo = map[string][]string{
    "title":  {},
    "year":   {},
    "track":  {},
    "genre":  {},
    "path":   {},
    "artist": {"performers"},
    "album":  {"albums"},
}

/*
Función que traduce un nodo, tratando cada caso según
sea una Condicion, OperacionLogica, Negacion u otro
*/
func (t *Traductor) Traduce(nodo Nodo) (string, []any, error) {
     switch n := nodo.(type) {
     case *Condicion:
	  if bool, err := verificaNodo(n); !bool {
	     return "", nil, err
	  }
	  campo, _ := camposPermitidos[n.Campo]
	  operador, _ := operadoresPermitidos[n.Operador]
     	  cadena := campo + " " + operador + " ?"
	  parametro := []any{}
	  if operador == "LIKE" {
	     parametro = append(parametro, "%" + n.Valor + "%")
	  } else {
	     parametro = append(parametro, n.Valor)
	  }
	  return cadena, parametro, nil

     case *OperacionLogica:
     	  cadA, paramA, err := t.Traduce(n.Izquierdo)
	  if err != nil {
	     return "", nil, err
	  }
	  cadB, paramB, err := t.Traduce(n.Derecho)
	  if err != nil {
	     return "", nil, err
	  }
	  var ol string
	  if n.Operador == "&" {
	     ol = "AND"
	  } else if n.Operador == "|"{
	    ol = "OR"
	  } else {
	    return "", nil, fmt.Errorf("Operador Lógico desconocido %s",
	    n.Operador)
	  }
	  cads := "(" + cadA + " " + ol + " " + cadB + ")"
	  params := slices.Concat(paramA, paramB)

	  return cads, params, nil
	  
     case *Negacion:
     	  cad, params, err := t.Traduce(n.Expresion)
	  if err != nil {
	     return "", nil, err
	  }
	  cadnot := "NOT (" + cad + ")"
	  return cadnot, params, nil

     default:
	return "", nil, fmt.Errorf("Tipo de Nodo desconocido: %T", n)
     
     }
     
     return "", nil, nil
}

/*
Función auxiliar que comprueba que un nodo
tenga un campo y operador válido
*/
func verificaNodo(n *Condicion) (bool, error) {
     if _, existe := camposPermitidos[n.Campo]; !existe {
     	return false, fmt.Errorf("Campo no permitido %s", n.Campo)
     }
     if _, existe := operadoresPermitidos[n.Operador]; !existe {
	return false, fmt.Errorf("Operador no permitido %s", n.Operador)
     }
     if coso, _ := operadoresPorCampo[n.Campo];
     	!slices.Contains(coso, n.Operador) {
	return false, fmt.Errorf("El operador %s no corresponde " +
	"con el campo %s", n.Operador, n.Campo)
     }
     if err := verificaTipo(n); err != nil {
     	return false, err
     }
     
     return true, nil
}

/*
Función que verifica que un valor esté correctamente
vinculado con el campo en una Condicion
*/
func verificaTipo(c *Condicion) error {
     tipo, ok := tiposPorCampo[c.Campo]
     if !ok {
     	return fmt.Errorf("Campo desconocido %s", c.Campo)
     }
     if c.TipoValor != tipo {
     	return fmt.Errorf("El tipo del valor no corresponde con el campo")
     }
     return nil
}

/*
Función que regresa los joins a partir de un Nodo
*/
func obtieneJoins(n Nodo) ([]string, error) {
     uniones := make(map[string]bool)

     var recorre func(Nodo) error

     recorre = func(n Nodo) error {
         switch nodo := n.(type) {
         case *Condicion:
             joins, ok := joinsPorCampo[nodo.Campo]
             if !ok {
                 return fmt.Errorf("No se conocen las uniones del campo %s", nodo.Campo)
             }

             for _, join := range joins {
                 uniones[join] = true
             } 

         case *OperacionLogica:
             if err := recorre(nodo.Izquierdo); err != nil {
                 return err
             }

             if err := recorre(nodo.Derecho); err != nil {
                 return err
             }

         case *Negacion:
             if err := recorre(nodo.Expresion); err != nil {
                 return err
             }

         default:
             return fmt.Errorf("Tipo de Nodo desconocido: %T", n)
         }

         return nil
     }

     if err := recorre(n); err != nil {
         return nil, err
     }

     resultado := []string{}

     for join := range uniones {
         resultado = append(resultado, join)
     }

     slices.Sort(resultado)
     return resultado, nil
}

/*
Función que construye una consulta SQL
a partir de un Nodo
*/
func (t *Traductor) ConstruyeConsulta(n Nodo) (string, []any, error) {
     where, parametros, err := t.Traduce(n)
     if err != nil {
         return "", nil, err
     }

     consulta := `SELECT
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
	      JOIN albums ON rolas.id_album = albums.id_album`

     consulta += " WHERE " + where
     return consulta, parametros, nil
}

