package parser

import (
	"strings"
	"testing"
)

func TestParseSimpleCommand(t *testing.T) {
	input := "GET mykey"
	
	cmd, errors := ParseCommand(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if cmd == nil {
		t.Fatal("ParseCommand returned nil")
	}
	
	if cmd.Command.Value != "GET" {
		t.Errorf("command name wrong. expected=GET, got=%s", cmd.Command.Value)
	}
	
	if len(cmd.Arguments) != 1 {
		t.Errorf("wrong number of arguments. expected=1, got=%d", len(cmd.Arguments))
	}
	
	arg, ok := cmd.Arguments[0].(*Identifier)
	if !ok {
		t.Errorf("argument is not Identifier. got=%T", cmd.Arguments[0])
	}
	
	if arg.Value != "mykey" {
		t.Errorf("argument value wrong. expected=mykey, got=%s", arg.Value)
	}
}

func TestParseSetCommandWithExpiration(t *testing.T) {
	input := `SET mykey "hello world" EX 60`
	
	cmd, errors := ParseCommand(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if cmd.Command.Value != "SET" {
		t.Errorf("command name wrong. expected=SET, got=%s", cmd.Command.Value)
	}
	
	if len(cmd.Arguments) != 4 {
		t.Errorf("wrong number of arguments. expected=4, got=%d", len(cmd.Arguments))
	}
	
	// Verificar clave
	key, ok := cmd.Arguments[0].(*Identifier)
	if !ok {
		t.Errorf("first argument is not Identifier. got=%T", cmd.Arguments[0])
	}
	if key.Value != "mykey" {
		t.Errorf("key wrong. expected=mykey, got=%s", key.Value)
	}
	
	// Verificar valor
	value, ok := cmd.Arguments[1].(*StringLiteral)
	if !ok {
		t.Errorf("second argument is not StringLiteral. got=%T", cmd.Arguments[1])
	}
	if value.Value != "hello world" {
		t.Errorf("value wrong. expected='hello world', got=%s", value.Value)
	}
	
	// Verificar opción EX
	ex, ok := cmd.Arguments[2].(*KeywordExpression)
	if !ok {
		t.Errorf("third argument is not KeywordExpression. got=%T", cmd.Arguments[2])
	}
	if ex.Value != "EX" {
		t.Errorf("option wrong. expected=EX, got=%s", ex.Value)
	}
	
	// Verificar tiempo de expiración
	expiry, ok := cmd.Arguments[3].(*IntegerLiteral)
	if !ok {
		t.Errorf("fourth argument is not IntegerLiteral. got=%T", cmd.Arguments[3])
	}
	if expiry.Value != 60 {
		t.Errorf("expiry wrong. expected=60, got=%d", expiry.Value)
	}
}

func TestParseZAddCommand(t *testing.T) {
	input := "ZADD myset 1.5 member1 2.0 member2"
	
	cmd, errors := ParseCommand(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if cmd.Command.Value != "ZADD" {
		t.Errorf("command name wrong. expected=ZADD, got=%s", cmd.Command.Value)
	}
	
	if len(cmd.Arguments) != 5 {
		t.Errorf("wrong number of arguments. expected=5, got=%d", len(cmd.Arguments))
	}
	
	// Verificar set name
	setName, ok := cmd.Arguments[0].(*Identifier)
	if !ok || setName.Value != "myset" {
		t.Errorf("set name wrong. expected=myset, got=%v", cmd.Arguments[0])
	}
	
	// Verificar primer score
	score1, ok := cmd.Arguments[1].(*FloatLiteral)
	if !ok || score1.Value != 1.5 {
		t.Errorf("first score wrong. expected=1.5, got=%v", cmd.Arguments[1])
	}
	
	// Verificar primer member
	member1, ok := cmd.Arguments[2].(*Identifier)
	if !ok || member1.Value != "member1" {
		t.Errorf("first member wrong. expected=member1, got=%v", cmd.Arguments[2])
	}
}

func TestParseZRangeWithOptions(t *testing.T) {
	input := "ZRANGE myset 0 -1 WITHSCORES"
	
	cmd, errors := ParseCommand(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if cmd.Command.Value != "ZRANGE" {
		t.Errorf("command name wrong. expected=ZRANGE, got=%s", cmd.Command.Value)
	}
	
	if len(cmd.Arguments) != 4 {
		t.Errorf("wrong number of arguments. expected=4, got=%d", len(cmd.Arguments))
	}
	
	// Verificar WITHSCORES option
	withScores, ok := cmd.Arguments[3].(*KeywordExpression)
	if !ok {
		t.Errorf("WITHSCORES is not KeywordExpression. got=%T", cmd.Arguments[3])
	}
	if withScores.Value != "WITHSCORES" {
		t.Errorf("option wrong. expected=WITHSCORES, got=%s", withScores.Value)
	}
}

func TestParseScanCommand(t *testing.T) {
	input := "SCAN 0 MATCH user:* COUNT 10"
	
	cmd, errors := ParseCommand(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if cmd.Command.Value != "SCAN" {
		t.Errorf("command name wrong. expected=SCAN, got=%s", cmd.Command.Value)
	}
	
	// El comando SCAN debería tener: cursor, MATCH, pattern, COUNT, count
	if len(cmd.Arguments) != 5 {
		t.Errorf("wrong number of arguments. expected=5, got=%d", len(cmd.Arguments))
	}
	
	// Verificar cursor
	cursor, ok := cmd.Arguments[0].(*IntegerLiteral)
	if !ok || cursor.Value != 0 {
		t.Errorf("cursor wrong. expected=0, got=%v", cmd.Arguments[0])
	}
	
	// Verificar MATCH keyword
	match, ok := cmd.Arguments[1].(*KeywordExpression)
	if !ok || match.Value != "MATCH" {
		t.Errorf("MATCH keyword wrong. expected=MATCH, got=%v", cmd.Arguments[1])
	}
	
	// Verificar pattern (debería ser parseado como PatternExpression)
	pattern, ok := cmd.Arguments[2].(*PatternExpression)
	if !ok {
		t.Errorf("pattern is not PatternExpression. got=%T", cmd.Arguments[2])
	}
	if pattern.Value != "user:*" {
		t.Errorf("pattern wrong. expected=user:*, got=%s", pattern.Value)
	}
}

func TestGetCommandInfo(t *testing.T) {
	input := `SET mykey "value" EX 60`
	
	cmd, errors := ParseCommand(input)
	if len(errors) != 0 {
		t.Fatalf("parser had errors: %v", errors)
	}
	
	info := GetCommandInfo(cmd)
	
	if info["command"] != "SET" {
		t.Errorf("command info wrong. expected=SET, got=%v", info["command"])
	}
	
	if info["arguments"] != 4 {
		t.Errorf("arguments count wrong. expected=4, got=%v", info["arguments"])
	}
	
	if info["has_key"] != true {
		t.Errorf("has_key should be true")
	}
	
	if info["has_value"] != true {
		t.Errorf("has_value should be true")
	}
	
	options, ok := info["options"].([]string)
	if !ok || len(options) != 1 || options[0] != "EX" {
		t.Errorf("options wrong. expected=[EX], got=%v", info["options"])
	}
}

func TestParseMultipleCommands(t *testing.T) {
	input := `SET key1 value1
GET key1
DEL key1`
	
	program, errors := ParseCommands(input)
	
	if len(errors) != 0 {
		t.Fatalf("parser had %d errors: %v", len(errors), errors)
	}
	
	if len(program.Statements) != 3 {
		t.Errorf("wrong number of statements. expected=3, got=%d", len(program.Statements))
	}
	
	// Verificar primer comando
	cmd1, ok := program.Statements[0].(*RedisCommand)
	if !ok {
		t.Errorf("first statement is not RedisCommand")
	}
	if cmd1.Command.Value != "SET" {
		t.Errorf("first command wrong. expected=SET, got=%s", cmd1.Command.Value)
	}
	
	// Verificar segundo comando
	cmd2, ok := program.Statements[1].(*RedisCommand)
	if !ok {
		t.Errorf("second statement is not RedisCommand")
	}
	if cmd2.Command.Value != "GET" {
		t.Errorf("second command wrong. expected=GET, got=%s", cmd2.Command.Value)
	}
	
	// Verificar tercer comando
	cmd3, ok := program.Statements[2].(*RedisCommand)
	if !ok {
		t.Errorf("third statement is not RedisCommand")
	}
	if cmd3.Command.Value != "DEL" {
		t.Errorf("third command wrong. expected=DEL, got=%s", cmd3.Command.Value)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123 key value", "expected command identifier"},
		{"", "expected command identifier"},
	}
	
	for _, tt := range tests {
		_, errors := ParseCommand(tt.input)
		
		if len(errors) == 0 {
			t.Errorf("expected parser error for input: %s", tt.input)
			continue
		}
		
		found := false
		for _, err := range errors {
			if strings.Contains(err, tt.expected) {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("expected error containing %q, got: %v", tt.expected, errors)
		}
	}
}

