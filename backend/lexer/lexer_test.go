package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `SET mykey "hello world" EX 60`
	
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "SET"},
		{IDENT, "mykey"},
		{STRING, "hello world"},
		{EX, "EX"},
		{INT, "60"},
		{EOF, ""},
	}
	
	l := New(input)
	
	for i, tt := range tests {
		tok := l.NextToken()
		
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestRedisCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "SET command with expiration",
			input: "SET key value EX 300",
			expected: []TokenType{IDENT, IDENT, IDENT, EX, INT, EOF},
		},
		{
			name:  "GET command",
			input: "GET mykey",
			expected: []TokenType{IDENT, IDENT, EOF},
		},
		{
			name:  "HSET command",
			input: "HSET myhash field1 value1",
			expected: []TokenType{IDENT, IDENT, IDENT, IDENT, EOF},
		},
		{
			name:  "ZADD with score",
			input: "ZADD myset 1.5 member1",
			expected: []TokenType{IDENT, IDENT, FLOAT, IDENT, EOF},
		},
		{
			name:  "ZRANGE with WITHSCORES",
			input: "ZRANGE myset 0 -1 WITHSCORES",
			expected: []TokenType{IDENT, IDENT, INT, INT, WITHSCORES, EOF},
		},
		{
			name:  "String with quotes",
			input: `SET key "hello world"`,
			expected: []TokenType{IDENT, IDENT, STRING, EOF},
		},
		{
			name:  "SCAN with MATCH pattern",
			input: "SCAN 0 MATCH user:* COUNT 10",
			expected: []TokenType{IDENT, INT, MATCH, IDENT, COLON, ASTERISK, COUNT, INT, EOF},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(tt.input)
			tokens := lexer.TokenizeCommand()
			
			if len(tokens) != len(tt.expected) {
				t.Fatalf("wrong number of tokens. expected=%d, got=%d",
					len(tt.expected), len(tokens))
			}
			
			for i, expectedType := range tt.expected {
				if tokens[i].Type != expectedType {
					t.Errorf("token[%d] wrong type. expected=%q, got=%q",
						i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	input := "SET key value"
	lexer := New(input)
	
	expectedPositions := []struct {
		literal  string
		position int
		line     int
		column   int
	}{
		{"SET", 0, 1, 1},
		{"key", 4, 1, 5},
		{"value", 8, 1, 9},
	}
	
	for i, expected := range expectedPositions {
		tok := lexer.NextToken()
		
		if tok.Literal != expected.literal {
			t.Errorf("token[%d] wrong literal. expected=%q, got=%q",
				i, expected.literal, tok.Literal)
		}
		
		if tok.Line != expected.line {
			t.Errorf("token[%d] wrong line. expected=%d, got=%d",
				i, expected.line, tok.Line)
		}
	}
}

func TestSpecialCharacters(t *testing.T) {
	input := "* ? [ ] ( ) , : | + -"
	
	expectedTypes := []TokenType{
		ASTERISK, QUESTION, BRACKET_L, BRACKET_R,
		PAREN_L, PAREN_R, COMMA, COLON, PIPE, PLUS, MINUS, EOF,
	}
	
	lexer := New(input)
	
	for i, expectedType := range expectedTypes {
		tok := lexer.NextToken()
		if tok.Type != expectedType {
			t.Errorf("token[%d] wrong type. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"123", INT, "123"},
		{"-456", INT, "-456"},
		{"3.14", FLOAT, "3.14"},
		{"-2.5", FLOAT, "-2.5"},
		{"0", INT, "0"},
	}
	
	for _, tt := range tests {
		lexer := New(tt.input)
		tok := lexer.NextToken()
		
		if tok.Type != tt.expectedType {
			t.Errorf("input %q: wrong type. expected=%q, got=%q",
				tt.input, tt.expectedType, tok.Type)
		}
		
		if tok.Literal != tt.expectedLit {
			t.Errorf("input %q: wrong literal. expected=%q, got=%q",
				tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

