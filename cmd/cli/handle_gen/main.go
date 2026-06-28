package main

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter seller first name: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	name := strings.TrimSpace(input)
	runes := []rune(name)

	limit := 4
	if len(runes) < limit {
		limit = len(runes)
	}

	prefix := string(runes[:limit])
	randomNum := rand.IntN(900) + 100

	handle := fmt.Sprintf("%s%d", prefix, randomNum)
	fmt.Println("Generated handle:", handle)
}
