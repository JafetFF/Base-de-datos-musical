package compilador

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