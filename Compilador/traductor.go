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
    "stage_name":  "persons.stage_name",
    "real_name":   "persons.real_name",
    "birth_date":  "persons.birth_date",
    "death_date":  "persons.death_date",
    "group_name":  "groups.name",
    "group_start": "groups.start_date",
    "group_end":   "groups.end_date",
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
	  parametro := []any{n.Valor}
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
     	  /*
     	  c, p, err := Traduce(n.Expresion)
	  if err != nil {
	     return "", nil, err
	  }
	  return */

     default:
     
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
     return true, nil
}

