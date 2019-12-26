package main

import (
	"fmt"
	"os"
)

type Notes struct {
	query []rune
}

func (n *Notes) Append(c rune) {
	n.query = append(n.query, c)
	fmt.Fprintf(os.Stderr, "query %s\r\n", string(n.query))
}

func (n *Notes) Backspace() {
	n.query = n.query[:len(n.query)-1]
	fmt.Fprintf(os.Stderr, "query %s\r\n", string(n.query))
}
