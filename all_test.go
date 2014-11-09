package main

import (
	"fmt"
	"testing"
)

func TestRuneWidth(t *testing.T) {
	fmt.Printf("%x %d\n", 'て', runeWidth('て'))
}
