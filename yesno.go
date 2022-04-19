package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func yesno(format string, args ...any) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(format+" [y/n]: ", args...)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("[ERR] read stdin: %v", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}
