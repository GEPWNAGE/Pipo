package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"./nonsentence"
)

func main() {
	ns, err := nonsentence.New("nonsentence.db")
	if err != nil {
		log.Fatal(err)
	}
	defer ns.Close()
	
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Ignore first word if it ends with a ':'
		var words = strings.Fields(scanner.Text())
		if (len(words) > 0) && strings.HasSuffix(words[0], ":") {
			words = words[1:]
		}
		if err := ns.Add(strings.Join(words, " ")); err != nil {
			log.Printf("Error while adding sentence: %v", err)
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
