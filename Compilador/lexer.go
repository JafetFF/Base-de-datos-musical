package compilador

import (
       "unicode"
       "fmt"
)

/*
El Lexer tiene una entrada que es la cadena recibida
y una posición para ir procesando cada parte de la entrada
*/
type Lexer struct {
     entrada  string
     posicion int
}

/*
Función auxiliar que regresa el siguiente token de la entrada
*/
func (l *Lexer) SigToken() (Token, error) {
     l.OmiteEspacios()

     if l.posicion >= len(l.entrada) {
        return Token{Tipo: EOF}, nil
     }

     var token Token

     switch l.entrada[l.posicion] {
     case '=':
         token = Token{Tipo: IGUAL, Valor: "="}

     case '!':
         if l.posicion+1 < len(l.entrada) && l.entrada[l.posicion+1] == '=' {
    	    token = Token{Tipo: DIFERENTE, Valor: "!="}
	    l.posicion++
	 } else {
    	    token = Token{Tipo: NOT, Valor: "!"}
	 }

     case '>':
         if l.posicion+1 < len(l.entrada) && l.entrada[l.posicion+1] == '=' {
    	    token = Token{Tipo: MAYOR_IGUAL, Valor: ">="}
	    l.posicion++
	 } else {
    	    token = Token{Tipo: MAYOR, Valor: ">"}
	 }

     case '<':
         if l.posicion+1 < len(l.entrada) && l.entrada[l.posicion+1] == '=' {
    	    token = Token{Tipo: MENOR_IGUAL, Valor: "<="}
	    l.posicion++
	 } else {
    	    token = Token{Tipo: MENOR, Valor: "<"}
	 }

     case '~':
         token = Token{Tipo: CONTIENE, Valor: "~"}

     case '&':
         token = Token{Tipo: AND, Valor: "&"}

     case '|':
         token = Token{Tipo: OR, Valor: "|"}

     case '(':
         token = Token{Tipo: PARENTESIS_IZQ, Valor: "("}
 
     case ')':
         token = Token{Tipo: PARENTESIS_DER, Valor: ")"}

     default:
         return Token{}, fmt.Errorf("Caracter no reconocido %c " +
	 "en la posición %d", l.entrada[l.posicion], l.posicion)	
     }
     l.posicion++

     return token, nil
}

/*
Función auxiliar que omite los espacios de la estrada
*/
func (l * Lexer) OmiteEspacios() {
     for ; l.posicion < len(l.entrada); l.posicion++ {
     	 if !unicode.IsSpace(rune(l.entrada[l.posicion])) {
	    return
	 }
     }
}