package compilador

import "testing"

/*
Prueba Unitaria para verificar que los campos
con el método SigToken() se leen correctamente
*/
func TestLeeCampos(t *testing.T) {
     lexer := Lexer{entrada: "title artist album"}

     esperados := []string{"title", "artist", "album"}

     for _, esperado := range esperados {
         token, err := lexer.SigToken()

         if err != nil {
             t.Fatalf("Ocurrió un error al intentar obtener el token %v",
	     err)
         }
 
         if token.Tipo != CAMPO {
             t.Errorf("Tipo incorrecto. Se esperaba CAMPO y se obtuvo %v",
	     token.Tipo)
         }

         if token.Valor != esperado {
             t.Errorf("Valor incorrecto. Se esperaba %q y se obtuvo %q",
	     esperado, token.Valor)
         }
     }

     token, err := lexer.SigToken()

     if err != nil {
         t.Fatalf("Ocurrió un error al obtener el siguiente token %v", err)
     }

     if token.Tipo != EOF {
         t.Errorf("Se esperaba EOF y se obtuvo %v", token.Tipo)
     }
}

/*
Prueba Unitaria para verificar que se leen correctamente
las cadenas pasadas al método SigToken()
*/
func TestLeeCadenas(t *testing.T) {
     lexer := Lexer{entrada: `"Where Do I?" "Care" "Su Hijo Dio"`}

     esperados := []string{
         "Where Do I?",
         "Care",
         "Su Hijo Dio",
     }

     for _, esperado := range esperados {
         token, err := lexer.SigToken()

         if err != nil {
             t.Fatalf("Ocurrió un error al leer el siguiente token %v", err)
         }

         if token.Tipo != CADENA {
             t.Errorf("Tipo incorrecto. Se esperaba CADENA y se obtuvo %v",
	     token.Tipo)
         }

         if token.Valor != esperado {
             t.Errorf("Valor incorrecto. Se esperaba %q y se obtuvo %q",
	     esperado, token.Valor)
         }
     }
}

/*
Prueba unitaria para verificar que se leen
correctamente los números
*/
func TestLeeNumeros(t *testing.T) {
     lexer := Lexer{entrada: "1500 2026 123"}

     esperados := []string{"1500", "2026", "123"}

     for _, esperado := range esperados {
         token, err := lexer.SigToken()
 
         if err != nil {
             t.Fatalf("Ocurrió un error al obtener el siguiente token %v",
	     err)
         }

         if token.Tipo != NUMERO {
             t.Errorf("Tipo incorrecto. Se esperaba NUMERO y se obtuvo %v",
	     token.Tipo)
         }

         if token.Valor != esperado {
             t.Errorf("Valor incorrecto. Se esperaba %q y se obtuvo %q",
	     esperado, token.Valor)
         }
     }
}

/*
Prueba unitaria para probar que se
leen correctamente los operadores
*/
func TestLeeOperadores(t *testing.T) {
    lexer := Lexer{entrada: "= != > < >= <="}

    esperados := []struct {
        tipo  TipoToken
        valor string
    }{
        {IGUAL, "="},
        {DIFERENTE, "!="},
        {MAYOR, ">"},
        {MENOR, "<"},
        {MAYOR_IGUAL, ">="},
        {MENOR_IGUAL, "<="},
    }

    for _, esperado := range esperados {
        token, err := lexer.SigToken()

        if err != nil {
            t.Fatalf("Error al leer el siguiente token %v", err)
        }
        if token.Tipo != esperado.tipo {
            t.Errorf("Tipo incorrecto, pues se esperaba %v y se obtuvo %v",
	    esperado.tipo, token.Tipo)
        }
        if token.Valor != esperado.valor {
            t.Errorf("Valor incorrecto, pues se esperaba %q y se obtuvo %q",
            esperado.valor, token.Valor)
        }
    }
}

/*
Prueba Unitaria que verifica que se lean
correctamente los operadores lógicos, así
también como los paréntesis
*/
func TestLeeLogicosYParentesis(t *testing.T) {
     lexer := Lexer{entrada: "& | ! ( )"}

     esperados := []struct {
         tipo  TipoToken
         valor string
     }{
         {AND, "&"},
         {OR, "|"},
         {NOT, "!"},
         {PARENTESIS_IZQ, "("},
         {PARENTESIS_DER, ")"},
     }

     for _, esperado := range esperados {
         token, err := lexer.SigToken()

         if err != nil {
             t.Fatalf("Error al obtener el siguiente token %v", err)
         }

         if token.Tipo != esperado.tipo {
             t.Errorf("Tipo incorrecto, pues se esperaba %v y se obtuvo %v",
             esperado.tipo, token.Tipo)
         }

         if token.Valor != esperado.valor {
             t.Errorf("Valor incorrecto, pues se esperaba %q y se obtuvo %q",
             esperado.valor, token.Valor)
         }
     }
}

/*
Prueba unitaria que comprueba que ante un
carácter desconocido se devuelve un error
*/ 
func TestLeeError(t *testing.T) {
     lexer := Lexer{entrada: "@"}

     _, err := lexer.SigToken()

     if err == nil {
         t.Error("Se esperaba un error por carácter desconocido")
     }
}

/*
Prueba unitaria que verifica que se
devuelve un error al recibir una cadena
sin cerrar con comillas 
*/
func TestCadenaSinCerrar(t *testing.T) {
     lexer := Lexer{entrada: `"Con Todo`}

     _, err := lexer.SigToken()

     if err == nil {
         t.Error("Se esperaba un error por falta de comillas")
     }
}

/*
Prueba Unitaria para verificar que al final
de una cadena se devuelve un token tipo EOF
(End Of File)
*/
func TestLeeEOF(t *testing.T) {
     lexer := Lexer{entrada: ""}

     token, err := lexer.SigToken()

     if err != nil {
         t.Fatalf("Error al obtener el siguiente token %v", err)
     }
 
     if token.Tipo != EOF {
         t.Errorf("Se esperaba EOF y se obtuvo %v", token.Tipo)
     }
}