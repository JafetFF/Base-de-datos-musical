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

// Constructor de Lexer
func NewLexer(entrada string) *Lexer {
     return &Lexer{
     	    entrada: entrada,
     }
}

/*
Función auxiliar que regresa el siguiente token de la entrada
*/
func (l *Lexer) SigToken() (Token, error) {
     l.OmiteEspacios()

     if l.posicion >= len(l.entrada) {
        return Token{Tipo: EOF}, nil
     }
     if unicode.IsLetter(rune(l.entrada[l.posicion])) {
     	return Token{Tipo: CAMPO, Valor: l.LeePalabra()}, nil
     }
     if unicode.IsDigit(rune(l.entrada[l.posicion])) {
     	return Token{Tipo: NUMERO, Valor: l.LeeNumero()}, nil
     }
     if l.entrada[l.posicion] == '"' {   
     	coso, err := l.LeeCadena()
	if err != nil {
	   return Token{}, err
	}
     	return Token{Tipo: CADENA, Valor: coso}, nil
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

/*
Función que lee una palabra de la cadena de entrada
*/
func (l *Lexer) LeePalabra() string {
     inicio := l.posicion
     for ; l.posicion < len(l.entrada); l.posicion++ {
     	 if !unicode.IsLetter(rune(l.entrada[l.posicion])) &&
	    l.entrada[l.posicion] != '_'{
	    break
	 }
     }
     
     return l.entrada[inicio:l.posicion]
}

/*
Función que lee una cadena que será la que
se busque en el campo correspondiente
*/
func (l *Lexer) LeeCadena() (string, error) {
     l.posicion++
     inicio := l.posicion
     for ; l.posicion < len(l.entrada); l.posicion++ {
     	 if l.entrada[l.posicion] == '"' {
	    break
	 }
     }
     if l.posicion >= len(l.entrada) {
     	return "", fmt.Errorf("Faltan comillas que cierren")
     }
     l.posicion++
     return l.entrada[inicio:l.posicion-1], nil
}

/*
Función que lee un número que será lo que corresponda a
campos con números como el año o la pista
*/
func (l *Lexer) LeeNumero() string {
     inicio := l.posicion
     for ; l.posicion < len(l.entrada); l.posicion++ {
     	 if !unicode.IsDigit(rune(l.entrada[l.posicion])) {
	    break
	 }
     }
     
     return l.entrada[inicio:l.posicion]
}    