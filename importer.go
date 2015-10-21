package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"./nonsentence"
)

const (
	IMPORT_BLOCK = 10000
)

func main() {
	ns, err := nonsentence.New("nonsentence.db")
	if err != nil {
		log.Fatal(err)
	}
	defer ns.Close()
	
	scanner := bufio.NewScanner(os.Stdin)
	sentences := make([]string, 0, IMPORT_BLOCK)
	for scanner.Scan() {
		// Ignore first word if it ends with a ':'
		var words = strings.Fields(scanner.Text())
		if (len(words) > 0) && strings.HasSuffix(words[0], ":") {
			words = words[1:]
		}
		sentences = append(sentences, strings.Join(words, " "))
		if len(sentences) == IMPORT_BLOCK {
			if err := ns.AddMultiple(sentences); err != nil {
				log.Printf("Error while adding sentences: %v", err)
			}
			sentences = make([]string, 0, IMPORT_BLOCK)
		}
	}
	if len(sentences) > 0 {
		if err := ns.AddMultiple(sentences); err != nil {
			log.Printf("Error while adding sentences: %v", err)
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
