package lexer

import (
	"strings"
)

// Lexer representa el analizador léxico
type Lexer struct {
	input        string
	position     int  // posición actual en input (apunta al carácter actual)
	readPosition int  // posición de lectura actual en input (después del carácter actual)
	ch           byte // carácter actual bajo examinación
	line         int  // línea actual
	column       int  // columna actual
}

// New crea un nuevo lexer
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar lee el siguiente carácter y avanza la posición en el input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL representa "EOF"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar devuelve el siguiente carácter sin avanzar la posición
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken escanea el input y devuelve el siguiente token
func (l *Lexer) NextToken() Token {
	var tok Token
	
	// Saltar espacios en blanco (excepto cuando son significativos)
	l.skipWhitespace()
	
	switch l.ch {
	case '*':
		tok = l.newToken(ASTERISK, l.ch)
	case '?':
		tok = l.newToken(QUESTION, l.ch)
	case '[':
		tok = l.newToken(BRACKET_L, l.ch)
	case ']':
		tok = l.newToken(BRACKET_R, l.ch)
	case '(':
		tok = l.newToken(PAREN_L, l.ch)
	case ')':
		tok = l.newToken(PAREN_R, l.ch)
	case ',':
		tok = l.newToken(COMMA, l.ch)
	case ':':
		tok = l.newToken(COLON, l.ch)
	case '|':
		tok = l.newToken(PIPE, l.ch)
	case '+':
		tok = l.newToken(PLUS, l.ch)
	case '-':
		// Podría ser un número negativo o solo un símbolo
		if isDigit(l.peekChar()) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			// Verificar si es un float
			if strings.Contains(tok.Literal, ".") {
				tok.Type = FLOAT
			}
			tok.Position = l.position - len(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column - len(tok.Literal)
			return tok
		}
		tok = l.newToken(MINUS, l.ch)
	case '"':
		tok.Type = STRING
		startPos := l.position
		tok.Literal = l.readString()
		tok.Position = startPos
		tok.Line = l.line
		tok.Column = l.column - len(tok.Literal) - 2
		return tok
	case '\'':
		tok.Type = STRING
		startPos := l.position
		tok.Literal = l.readSingleQuoteString()
		tok.Position = startPos
		tok.Line = l.line
		tok.Column = l.column - len(tok.Literal) - 2
		return tok
	case '\n':
		tok = l.newToken(NEWLINE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		tok.Position = l.position
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(strings.ToUpper(tok.Literal))
			tok.Position = l.position - len(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column - len(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			// Verificar si es un float
			if strings.Contains(tok.Literal, ".") {
				tok.Type = FLOAT
			}
			tok.Position = l.position - len(tok.Literal)
			tok.Line = l.line
			tok.Column = l.column - len(tok.Literal)
			return tok
		} else {
			tok = l.newToken(ILLEGAL, l.ch)
		}
	}
	
	l.readChar()
	return tok
}

// newToken crea un nuevo token con el tipo y carácter dados
func (l *Lexer) newToken(tokenType TokenType, ch byte) Token {
	return Token{
		Type:     tokenType,
		Literal:  string(ch),
		Position: l.position,
		Line:     l.line,
		Column:   l.column,
	}
}

// readIdentifier lee un identificador (comando Redis o argumento)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '-' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber lee un número (entero o flotante)
func (l *Lexer) readNumber() string {
	position := l.position
	
	// Manejar signo negativo
	if l.ch == '-' {
		l.readChar()
	}
	
	// Leer parte entera
	for isDigit(l.ch) {
		l.readChar()
	}
	
	// Verificar si hay parte decimal
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consumir el '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	
	return l.input[position:l.position]
}

// readString lee una cadena entre comillas dobles
func (l *Lexer) readString() string {
	position := l.position + 1 // saltar la comilla inicial
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		// Manejar escape sequences
		if l.ch == '\\' {
			l.readChar() // saltar el carácter de escape
			if l.ch != 0 {
				l.readChar() // saltar el carácter escapado
			}
		}
	}
	result := l.input[position:l.position]
	// Avanzar más allá de la comilla de cierre
	if l.ch == '"' {
		l.readChar()
	}
	return result
}

// readSingleQuoteString lee una cadena entre comillas simples
func (l *Lexer) readSingleQuoteString() string {
	position := l.position + 1 // saltar la comilla inicial
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
		// Manejar escape sequences
		if l.ch == '\\' {
			l.readChar() // saltar el carácter de escape
			if l.ch != 0 {
				l.readChar() // saltar el carácter escapado
			}
		}
	}
	result := l.input[position:l.position]
	// Avanzar más allá de la comilla de cierre
	if l.ch == '\'' {
		l.readChar()
	}
	return result
}

// skipWhitespace salta espacios en blanco excepto nuevas líneas
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter verifica si el carácter es una letra
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit verifica si el carácter es un dígito
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// TokenizeCommand tokeniza un comando Redis completo
func (l *Lexer) TokenizeCommand() []Token {
	var tokens []Token
	
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			tokens = append(tokens, tok)
			break
		}
		// Filtrar espacios en blanco para simplificar el análisis
		if tok.Type != SPACE {
			tokens = append(tokens, tok)
		}
	}
	
	return tokens
}

// GetAllTokens devuelve todos los tokens del input
func GetAllTokens(input string) []Token {
	lexer := New(input)
	return lexer.TokenizeCommand()
}

