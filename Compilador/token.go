package compilador

// El TipoToken es la categoría o el tipo de token
type TipoToken int
const (
     CAMPO TipoToken = iota
     CADENA
     NUMERO
     IGUAL
     DIFERENTE
     CONTIENE
     MAYOR
     MAYOR_IGUAL
     MENOR
     MENOR_IGUAL
     AND
     OR
     NOT
     PARENTESIS_IZQ
     PARENTESIS_DER
     EOF
)

// Token representa un elemento léxico que lexer identifica
type Token struct {
     Tipo  TipoToken
     Valor string
}