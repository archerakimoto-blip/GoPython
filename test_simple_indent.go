package main

import (
	"fmt"

	"github.com/go-py/go-python/pkg/lexer"
)

func main() {
	input := `class Animal:
    def speak(self):
        return "Animal sound"

class Dog(Animal):
    def speak(self):
        return "Woof!"`

	fmt.Println("Input code:")
	fmt.Println(input)

	fmt.Println("\n--- Lexer tokens ---")
	l := lexer.New(input)
	for i := 0; i < 50; i++ {
		tok := l.NextToken()
		fmt.Printf("Token %2d: Type: %-10s Literal: %q\n", i, tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}