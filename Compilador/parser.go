package compilador

import "fmt"

type Parser struct {
     lex         *Lexer
     tokenActual Token
}

// Constructor del Parser
func NewParser(lexer *Lexer) (*Parser, error) {
     token, err := lexer.SigToken()
     if err != nil {
     	return nil, err
     }
     return &Parser{lex: lexer, tokenActual: token}, err
}

// Función que actualiza el token del Parser y devuelve error si hubo
func (p *Parser) AvanzaToken() error {
     token, err := p.lex.SigToken()
     if err != nil {
     	return err
     }
     p.tokenActual = token
     return nil
}

// Función principal que parsea una entrada completa
func (p *Parser) Parsea() (Nodo, error) {
     nodo, err := p.parseaExpresion()
     if err != nil {
     	return nil, err
     }
     if p.tokenActual.Tipo != EOF {
     	return nil, fmt.Errorf("Error al analizar la entrada " +
	"pues queda parte de la entrada sin analizar")
     }
     return nodo, nil
}

// Función que parsea una condición
func (p *Parser) parseaCondicion() (*Condicion, error) {
     c := Condicion{}
     switch p.tokenActual.Tipo {
     case CAMPO:
     	  c.Campo = p.tokenActual.Valor
	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }

     default:
	return nil, fmt.Errorf("Se esperaba un campo")
     }
     
     switch p.tokenActual.Tipo {
     case IGUAL, DIFERENTE, CONTIENE, MAYOR,
          MENOR, MAYOR_IGUAL, MENOR_IGUAL:
	  c.Operador = p.tokenActual.Valor
	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }

     default:
	return nil, fmt.Errorf("Se esperaba un operador")
     }

     switch p.tokenActual.Tipo {
     case CADENA:
     	  c.Valor = p.tokenActual.Valor
	  c.TipoValor = CADENA
	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }

     case NUMERO:
     	  c.Valor = p.tokenActual.Valor
	  c.TipoValor = NUMERO
	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }

     default:
	return nil, fmt.Errorf("Se esperaba una cadena o número")
     }
     
     return &c, nil
}

// Función que parsea un factor recursivamente
func (p *Parser) parseaFactor() (Nodo, error) {
     switch p.tokenActual.Tipo {
     case NOT:
     	  not := Negacion{}
     	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }
	  expresion, err := p.parseaFactor()
	  if err != nil {
	     return nil, err
	  }
	  not.Expresion = expresion
	  return &not, nil

     case PARENTESIS_IZQ:
     	  err := p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }
	  
	  expresion, err := p.parseaExpresion()
	  if err != nil {
	     return nil, err
	  }
	  
	  if p.tokenActual.Tipo != PARENTESIS_DER {
	     return nil, fmt.Errorf("Debes cerrar los paréntesis")
	  }
	  err = p.AvanzaToken()
	  if err != nil {
	     return nil, err
	  }
	  return expresion, nil

     default:
	return p.parseaCondicion()
     }
}

// Función que parsea términos
func (p *Parser) parseaTermino() (Nodo, error) {
     izq, err := p.parseaFactor()
     if err != nil {
     	return nil, err
     }
     for p.tokenActual.Tipo == AND {     	 
     	 err := p.AvanzaToken()
	 if err != nil {
	    return nil, err
	 }
	 der, err := p.parseaFactor()
         if err != nil {
	    return nil, err
	 }
	 izq = &OperacionLogica{
	     Operador:  "&",
	     Izquierdo: izq,
	     Derecho:   der,
	 }
     }
     return izq, nil
}

// Funión que parsea una expresión completa
func (p *Parser) parseaExpresion() (Nodo, error) {
     izq, err := p.parseaTermino()
     if err != nil {
     	return nil, err
     }
     for p.tokenActual.Tipo == OR {     	 
     	 err := p.AvanzaToken()
	 if err != nil {
	    return nil, err
	 }
	 der, err := p.parseaTermino()
         if err != nil {
	    return nil, err
	 }
	 izq = &OperacionLogica{
	     Operador:  "|",
	     Izquierdo: izq,
	     Derecho:   der,
	 }
     }
     return izq, nil
}