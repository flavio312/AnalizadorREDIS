package parser

import (
	"fmt"
	"strconv"
	"redis-analyzer-api/lexer"
)

// Parser representa el analizador sintáctico
type Parser struct {
	lexer *lexer.Lexer
	
	curToken  lexer.Token
	peekToken lexer.Token
	
	errors []string
}

// New crea un nuevo parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}
	
	// Leer dos tokens para que curToken y peekToken estén configurados
	p.nextToken()
	p.nextToken()
	
	return p
}

// nextToken avanza los tokens
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// Errors devuelve los errores de parsing
func (p *Parser) Errors() []string {
	return p.errors
}

// addError añade un error al parser
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("Parser error at line %d, column %d: %s", 
		p.curToken.Line, p.curToken.Column, msg))
}

// ParseProgram parsea el programa completo
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}
	
	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	
	return program
}

// parseStatement parsea una declaración
func (p *Parser) parseStatement() Statement {
	// Saltar nuevas líneas
	for p.curToken.Type == lexer.NEWLINE {
		p.nextToken()
	}
	
	// Si llegamos al final, retornar nil
	if p.curToken.Type == lexer.EOF {
		return nil
	}
	
	// En Redis, todas las declaraciones son comandos
	return p.parseRedisCommand()
}

// parseRedisCommand parsea un comando Redis
func (p *Parser) parseRedisCommand() *RedisCommand {
	if p.curToken.Type != lexer.IDENT {
		p.addError(fmt.Sprintf("expected command identifier, got %s", p.curToken.Type))
		return nil
	}
	
	cmd := &RedisCommand{}
	cmd.Command = &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	
	// Parsear argumentos
	cmd.Arguments = []Expression{}
	
	for p.peekToken.Type != lexer.EOF && p.peekToken.Type != lexer.NEWLINE {
		p.nextToken()
		arg := p.parseExpression()
		if arg != nil {
			cmd.Arguments = append(cmd.Arguments, arg)
		}
	}
	
	return cmd
}

// parseExpression parsea una expresión
func (p *Parser) parseExpression() Expression {
	switch p.curToken.Type {
	case lexer.IDENT:
		// Verificar si es parte de un patrón (ej: user:*)
		if p.peekToken.Type == lexer.COLON {
			return p.parsePatternExpression()
		}
		return p.parseIdentifier()
	case lexer.STRING:
		return p.parseStringLiteral()
	case lexer.INT:
		return p.parseIntegerLiteral()
	case lexer.FLOAT:
		return p.parseFloatLiteral()
	case lexer.EX, lexer.PX, lexer.NX, lexer.XX, lexer.WITHSCORES, 
		 lexer.LIMIT, lexer.COUNT, lexer.MATCH, lexer.TYPE:
		return p.parseKeywordExpression()
	case lexer.ASTERISK, lexer.QUESTION:
		return p.parsePatternExpression()
	case lexer.BRACKET_L:
		return p.parseRangeExpression()
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

// parseIdentifier parsea un identificador
func (p *Parser) parseIdentifier() *Identifier {
	return &Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseStringLiteral parsea una cadena literal
func (p *Parser) parseStringLiteral() *StringLiteral {
	return &StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseIntegerLiteral parsea un entero literal
func (p *Parser) parseIntegerLiteral() *IntegerLiteral {
	lit := &IntegerLiteral{Token: p.curToken}
	
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}
	
	lit.Value = value
	return lit
}

// parseFloatLiteral parsea un flotante literal
func (p *Parser) parseFloatLiteral() *FloatLiteral {
	lit := &FloatLiteral{Token: p.curToken}
	
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addError(fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}
	
	lit.Value = value
	return lit
}

// parseKeywordExpression parsea palabras clave de Redis
func (p *Parser) parseKeywordExpression() *KeywordExpression {
	return &KeywordExpression{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parsePatternExpression parsea patrones con wildcards
func (p *Parser) parsePatternExpression() *PatternExpression {
	pattern := p.curToken.Literal
	
	// Si hay más símbolos de patrón consecutivos, combinarlos
	for p.peekToken.Type == lexer.ASTERISK || p.peekToken.Type == lexer.QUESTION ||
		p.peekToken.Type == lexer.IDENT || p.peekToken.Type == lexer.COLON {
		p.nextToken()
		pattern += p.curToken.Literal
	}
	
	return &PatternExpression{
		Token: p.curToken,
		Value: pattern,
	}
}

// parseRangeExpression parsea expresiones de rango [start, end]
func (p *Parser) parseRangeExpression() *RangeExpression {
	if p.curToken.Type != lexer.BRACKET_L {
		p.addError("expected '['")
		return nil
	}
	
	p.nextToken() // consumir '['
	
	start := p.parseExpression()
	if start == nil {
		return nil
	}
	
	if p.peekToken.Type != lexer.COMMA {
		p.addError("expected ',' in range expression")
		return nil
	}
	p.nextToken() // consumir ','
	p.nextToken() // ir al siguiente elemento
	
	end := p.parseExpression()
	if end == nil {
		return nil
	}
	
	if p.peekToken.Type != lexer.BRACKET_R {
		p.addError("expected ']'")
		return nil
	}
	p.nextToken() // consumir ']'
	
	return &RangeExpression{
		Start: start,
		End:   end,
	}
}

// ParseCommand parsea un solo comando Redis desde una cadena
func ParseCommand(input string) (*RedisCommand, []string) {
	l := lexer.New(input)
	p := New(l)
	
	cmd := p.parseRedisCommand()
	return cmd, p.Errors()
}

// ParseCommands parsea múltiples comandos Redis
func ParseCommands(input string) (*Program, []string) {
	l := lexer.New(input)
	p := New(l)
	
	program := p.ParseProgram()
	return program, p.Errors()
}

// GetCommandInfo extrae información básica de un comando
func GetCommandInfo(cmd *RedisCommand) map[string]interface{} {
	info := map[string]interface{}{
		"command":    cmd.Command.Value,
		"arguments":  len(cmd.Arguments),
		"arg_types":  []string{},
		"options":    []string{},
		"has_key":    false,
		"has_value":  false,
	}
	
	argTypes := []string{}
	options := []string{}
	
	for i, arg := range cmd.Arguments {
		argType := arg.Type()
		argTypes = append(argTypes, argType)
		
		// Detectar si es una clave (primer argumento en la mayoría de comandos)
		if i == 0 && (argType == "Identifier" || argType == "StringLiteral") {
			info["has_key"] = true
		}
		
		// Detectar valores
		if argType == "StringLiteral" || argType == "IntegerLiteral" || argType == "FloatLiteral" {
			info["has_value"] = true
		}
		
		// Detectar opciones
		if argType == "KeywordExpression" {
			options = append(options, arg.String())
		}
	}
	
	info["arg_types"] = argTypes
	info["options"] = options
	
	return info
}

