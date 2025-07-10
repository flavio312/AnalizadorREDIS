package parser

import (
	"fmt"
	"redis-analyzer-api/lexer"
)

// Node representa un nodo en el AST
type Node interface {
	String() string
	Type() string
}

// Statement representa una declaración en Redis
type Statement interface {
	Node
	statementNode()
}

// Expression representa una expresión en Redis
type Expression interface {
	Node
	expressionNode()
}

// RedisCommand representa un comando Redis completo
type RedisCommand struct {
	Command   *Identifier
	Arguments []Expression
}

func (rc *RedisCommand) statementNode() {}
func (rc *RedisCommand) String() string {
	args := ""
	for i, arg := range rc.Arguments {
		if i > 0 {
			args += " "
		}
		args += arg.String()
	}
	return fmt.Sprintf("%s %s", rc.Command.String(), args)
}
func (rc *RedisCommand) Type() string { return "RedisCommand" }

// Identifier representa un identificador (comando o clave)
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }
func (i *Identifier) Type() string    { return "Identifier" }

// StringLiteral representa una cadena de texto
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string  { return fmt.Sprintf("\"%s\"", sl.Value) }
func (sl *StringLiteral) Type() string    { return "StringLiteral" }

// IntegerLiteral representa un número entero
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) String() string  { return fmt.Sprintf("%d", il.Value) }
func (il *IntegerLiteral) Type() string    { return "IntegerLiteral" }

// FloatLiteral representa un número flotante
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}
func (fl *FloatLiteral) String() string  { return fmt.Sprintf("%f", fl.Value) }
func (fl *FloatLiteral) Type() string    { return "FloatLiteral" }

// KeywordExpression representa palabras clave de Redis como EX, PX, NX, etc.
type KeywordExpression struct {
	Token lexer.Token
	Value string
}

func (ke *KeywordExpression) expressionNode() {}
func (ke *KeywordExpression) String() string  { return ke.Value }
func (ke *KeywordExpression) Type() string    { return "KeywordExpression" }

// PatternExpression representa patrones con wildcards
type PatternExpression struct {
	Token lexer.Token
	Value string
}

func (pe *PatternExpression) expressionNode() {}
func (pe *PatternExpression) String() string  { return pe.Value }
func (pe *PatternExpression) Type() string    { return "PatternExpression" }

// RangeExpression representa rangos como [0, -1]
type RangeExpression struct {
	Start Expression
	End   Expression
}

func (re *RangeExpression) expressionNode() {}
func (re *RangeExpression) String() string {
	return fmt.Sprintf("[%s, %s]", re.Start.String(), re.End.String())
}
func (re *RangeExpression) Type() string { return "RangeExpression" }

// OptionExpression representa opciones de comandos con sus valores
type OptionExpression struct {
	Option Expression
	Value  Expression // puede ser nil si la opción no tiene valor
}

func (oe *OptionExpression) expressionNode() {}
func (oe *OptionExpression) String() string {
	if oe.Value != nil {
		return fmt.Sprintf("%s %s", oe.Option.String(), oe.Value.String())
	}
	return oe.Option.String()
}
func (oe *OptionExpression) Type() string { return "OptionExpression" }

// Program representa el programa completo (puede contener múltiples comandos)
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	result := ""
	for i, stmt := range p.Statements {
		if i > 0 {
			result += "\n"
		}
		result += stmt.String()
	}
	return result
}
func (p *Program) Type() string { return "Program" }

