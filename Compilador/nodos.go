package compilador

// Representa un Nodo del Árbol de Sintaxis Abstracta (ASA)
type Nodo interface {
     esNodo()
}

type Condicion struct {
     Campo    string
     Operador string
     Valor    string
}

type OperacionLogica struct {
     Operador  string
     Izquierdo Nodo
     Derecho   Nodo
}

type Negacion struct {
     Expresion Nodo
}

func (c Condicion) esNodo() {}
func (ol OperacionLogica) esNodo() {}
func (n Negacion) esNodo() {}
