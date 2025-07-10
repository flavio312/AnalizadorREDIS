package lexer

import "fmt"

// TokenType representa el tipo de token
type TokenType int

const (
	// Tipos especiales
	ILLEGAL TokenType = iota
	EOF
	
	// Identificadores y literales
	IDENT   // comandos Redis como SET, GET, HSET, etc.
	STRING  // cadenas de texto
	INT     // números enteros
	FLOAT   // números flotantes
	
	// Delimitadores
	SPACE     // espacios
	NEWLINE   // nueva línea
	
	// Símbolos especiales de Redis
	ASTERISK  // * (usado en patrones)
	QUESTION  // ? (usado en patrones)
	BRACKET_L // [ (usado en rangos)
	BRACKET_R // ] (usado en rangos)
	PAREN_L   // ( (usado en algunos comandos)
	PAREN_R   // ) (usado en algunos comandos)
	COMMA     // , (separador en algunos comandos)
	COLON     // : (usado en algunos comandos)
	PIPE      // | (usado en algunos comandos)
	PLUS      // + (usado en algunos comandos)
	MINUS     // - (usado en algunos comandos)
	
	// Palabras clave especiales de Redis
	EX        // EX (expiration)
	PX        // PX (expiration in milliseconds)
	NX        // NX (not exists)
	XX        // XX (exists)
	WITHSCORES // WITHSCORES
	LIMIT     // LIMIT
	COUNT     // COUNT
	MATCH     // MATCH
	TYPE      // TYPE
)

// Token representa un token individual
type Token struct {
	Type     TokenType
	Literal  string
	Position int
	Line     int
	Column   int
}

// String devuelve una representación en string del token
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Pos: %d, Line: %d, Col: %d}", 
		t.Type.String(), t.Literal, t.Position, t.Line, t.Column)
}

// String devuelve el nombre del tipo de token
func (tt TokenType) String() string {
	switch tt {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case STRING:
		return "STRING"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case SPACE:
		return "SPACE"
	case NEWLINE:
		return "NEWLINE"
	case ASTERISK:
		return "ASTERISK"
	case QUESTION:
		return "QUESTION"
	case BRACKET_L:
		return "BRACKET_L"
	case BRACKET_R:
		return "BRACKET_R"
	case PAREN_L:
		return "PAREN_L"
	case PAREN_R:
		return "PAREN_R"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case PIPE:
		return "PIPE"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case EX:
		return "EX"
	case PX:
		return "PX"
	case NX:
		return "NX"
	case XX:
		return "XX"
	case WITHSCORES:
		return "WITHSCORES"
	case LIMIT:
		return "LIMIT"
	case COUNT:
		return "COUNT"
	case MATCH:
		return "MATCH"
	case TYPE:
		return "TYPE"
	default:
		return "UNKNOWN"
	}
}

// keywords es un mapa de palabras clave de Redis
var keywords = map[string]TokenType{
	"EX":         EX,
	"PX":         PX,
	"NX":         NX,
	"XX":         XX,
	"WITHSCORES": WITHSCORES,
	"LIMIT":      LIMIT,
	"COUNT":      COUNT,
	"MATCH":      MATCH,
	"TYPE":       TYPE,
}

// LookupIdent verifica si un identificador es una palabra clave
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

