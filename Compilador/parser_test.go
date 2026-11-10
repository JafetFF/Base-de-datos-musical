package compilador

import "testing"

/*
Prueba unitaria para comprobar que se
parsea correctamente una condición
*/
func TestParseaCondicion(t *testing.T) {
     lexer := Lexer{entrada: `titulo = "name"`}
     parser, err := NewParser(&lexer)
     if err != nil {
     	t.Fatal(err)
     }
     con, err := parser.parseaCondicion()
     if err != nil {
     	t.Fatal(err)
     }
     if con.Campo != "titulo" {
	t.Errorf("Se esperaba titulo, y se obtuvo %s", con.Campo)
     }
     if con.Operador != "=" {
	t.Errorf("Se esperaba '=', y se obtuvo %s", con.Operador)
     }
     if con.Valor != "name" {
	t.Errorf("Se esperaba 'name', y se obtuvo %s", con.Valor)
     }
}

/*
Prueba unitaria para verificar que se parsea
correctamente una condición númerica
*/
func TestParseaCondicionNumerica(t *testing.T) {
    lexer := Lexer{entrada: `year >= 2007`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    con, err := parser.parseaCondicion()
    if err != nil {
        t.Fatal(err)
    }

    if con.Campo != "year" {
        t.Errorf("Se esperaba year, y se obtuvo %s", con.Campo)
    }

    if con.Operador != ">=" {
        t.Errorf("Se esperaba >=, y se obtuvo %s", con.Operador)
    }

    if con.Valor != "2007" {
        t.Errorf("Se esperaba 2007, y se obtuvo %s", con.Valor)
    }
}

/*
Prueba unitaria que verifica que se
obtenga un error como falta del campo
*/ 
func TestParseaCondicionSinCampo(t *testing.T) {
    lexer := Lexer{entrada: `= "name"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    _, err = parser.parseaCondicion()
    if err == nil {
        t.Error("Se esperaba un error por falta de campo")
    }
}

/*
Prueba unitaria que verifica que se
obtenga un error como falta del operador
*/ 
func TestParseaCondicionSinOperador(t *testing.T) {
    lexer := Lexer{entrada: `titulo "name"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    _, err = parser.parseaCondicion()
    if err == nil {
        t.Error("Se esperaba un error por falta de operador")
    }
}

/*
Prueba unitaria que verifica que se
obtenga un error como falta del valor
*/ 
func TestParseaCondicionSinValor(t *testing.T) {
    lexer := Lexer{entrada: `titulo =`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    _, err = parser.parseaCondicion()
    if err == nil {
        t.Error("Se esperaba un error por falta de valor")
    }
}

/*
Prueba unitaria para probar el caso más simple de parseaFactor()
*/
func TestParseaFactorCondicion(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "name"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }
    nodo, err := parser.parseaFactor()
    if err != nil {
        t.Fatal(err)
    }

    con, ok := nodo.(*Condicion)
    if !ok {
        t.Fatal("Se esperaba un nodo de tipo Condicion")
    }
    if con.Campo != "titulo" {
        t.Errorf("Se esperaba titulo, se obtuvo %s", con.Campo)
    }
}

/*
Prueba unitaria del caso de Negacion de parseaFactor()
*/
func TestParseaFactorNegacion(t *testing.T) {
    lexer := Lexer{entrada: `!titulo = "name"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }
    nodo, err := parser.parseaFactor()
    if err != nil {
        t.Fatal(err)
    }

    neg, ok := nodo.(*Negacion)
    if !ok {
        t.Fatal("Se esperaba un nodo de tipo Negacion")
    }
    con, ok := neg.Expresion.(*Condicion)
    if !ok {
        t.Fatal("Se esperaba una Condicion dentro de Negacion")
    }
    if con.Campo != "titulo" {
        t.Errorf("Se esperaba titulo, se obtuvo %s", con.Campo)
    }
}

/*
Prueba unitaria del caso de paréntesis de parseaFactor()
*/
func TestParseaFactorParentesis(t *testing.T) {
    lexer := Lexer{entrada: `(titulo = "name")`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.parseaFactor()
    if err != nil {
        t.Fatal(err)
    }
    con, ok := nodo.(*Condicion)
    if !ok {
        t.Fatal("Se esperaba un nodo de tipo Condicion")
    }

    if con.Campo != "titulo" {
        t.Errorf("Se esperaba titulo, se obtuvo %s", con.Campo)
    }
}

/*
Prueba unitaria de parseaTermino
*/
func TestParseaTermino(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "love" & genero = "Pop"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.parseaTermino()
    if err != nil {
        t.Fatal(err)
    }

    ol, ok := nodo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba un nodo de tipo OperacionLogica")
    }

    if ol.Operador != "&" {
        t.Errorf("Se esperaba &, y se obtuvo %s", ol.Operador)
    }

    if _, ok := ol.Izquierdo.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado izquierdo")
    }

    if _, ok := ol.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado derecho")
    }
}

/*
Prueba unitaria con más términos para parseaTermino()
*/
func TestParseaTerminoMultiple(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "luna" & genero = "Rock" & year = 2023`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.parseaTermino()
    if err != nil {
        t.Fatal(err)
    }

    ol, ok := nodo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba una OperacionLogica")
    }

    if ol.Operador != "&" {
        t.Errorf("Se esperaba &, se obtuvo %s", ol.Operador)
    }

    izq, ok := ol.Izquierdo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba una OperacionLogica en el lado izquierdo")
    }

    if izq.Operador != "&" {
        t.Errorf("Se esperaba &, se obtuvo %s", izq.Operador)
    }

    if _, ok := izq.Izquierdo.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado izquierdo")
    }

    if _, ok := izq.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado derecho")
    }

    if _, ok := ol.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado derecho")
    }
}

// Prueba unitaria para parseaExpresion()
func TestParseaExpresion(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "Day" | genero != "Rock"`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.parseaExpresion()
    if err != nil {
        t.Fatal(err)
    }

    ol, ok := nodo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba un nodo de tipo OperacionLogica")
    }

    if ol.Operador != "|" {
        t.Errorf("Se esperaba |, se obtuvo %s", ol.Operador)
    }

    if _, ok := ol.Izquierdo.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el nodo izquierdo")
    }

    if _, ok := ol.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el nodo derecho")
    }
}

/*
Prueba unitaria para verificar que se interpretan
correctamente los operadores lógicos
*/
func TestParseaPrecedencia(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "Café" & genero = "Rock" | year != 2000`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.parseaExpresion()
    if err != nil {
        t.Fatal(err)
    }

    or, ok := nodo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba una OperacionLogica")
    }

    if or.Operador != "|" {
        t.Errorf("Se esperaba |, se obtuvo %s", or.Operador)
    }

    and, ok := or.Izquierdo.(*OperacionLogica)
    if !ok {
        t.Fatal("Se esperaba una OperacionLogica en el lado izquierdo")
    }

    if and.Operador != "&" {
        t.Errorf("Se esperaba &, se obtuvo %s", and.Operador)
    }

    if _, ok := and.Izquierdo.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado izquierdo del AND")
    }

    if _, ok := and.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado derecho del AND")
    }

    if _, ok := or.Derecho.(*Condicion); !ok {
        t.Error("Se esperaba una Condicion en el lado derecho del OR")
    }
}

/*
Prueba unitaria para el método principal
*/
func TestParsea(t *testing.T) {
    lexer := Lexer{entrada: `(titulo != "Rain" & genero = "Rock") | year = 1908`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    nodo, err := parser.Parsea()
    if err != nil {
        t.Fatal(err)
    }

    if nodo == nil {
        t.Fatal("Se esperaba un árbol de sintaxis abstracta")
    }
}

/*
Prueba unitaria para las entradas inválidas
*/
func TestParseaEntradaInvalida(t *testing.T) {
    lexer := Lexer{entrada: `titulo = "nombre" year = 2006`}
    parser, err := NewParser(&lexer)
    if err != nil {
        t.Fatal(err)
    }

    _, err = parser.Parsea()
    if err == nil {
        t.Error("Se esperaba un error por tokens sobrantes")
    }
}