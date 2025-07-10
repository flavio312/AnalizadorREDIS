package main

import (
	"fmt"
	"strings"
	"redis-analyzer-api/lexer"
)

func main() {
	// Test case que está fallando
	input := `SET mykey "hello world" EX 60`
	
	fmt.Printf("Input: %s\n", input)
	fmt.Println("Tokens:")
	
	l := lexer.New(input)
	for {
		tok := l.NextToken()
		fmt.Printf("  %s\n", tok.String())
		if tok.Type == lexer.EOF {
			break
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	
	// Test case con comillas
	input2 := `SET key "hello world"`
	fmt.Printf("Input: %s\n", input2)
	fmt.Println("Tokens:")
	
	tokens := lexer.GetAllTokens(input2)
	for _, tok := range tokens {
		fmt.Printf("  %s\n", tok.String())
	}
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	
	// Test case con SCAN
	input3 := "SCAN 0 MATCH user:* COUNT 10"
	fmt.Printf("Input: %s\n", input3)
	fmt.Println("Tokens:")
	
	tokens3 := lexer.GetAllTokens(input3)
	for _, tok := range tokens3 {
		fmt.Printf("  %s\n", tok.String())
	}
}

