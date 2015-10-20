package main

import (
	"log"
	"./nonsentence"
)

func main() {
	log.Printf("Running Pipo");
	
	ns, err := nonsentence.New("nonsentence.db")
	if err != nil {
		log.Fatal(err)
	}
	defer ns.Close()

// 	ns.Add("De kat krabt de krullen van de trap.")
// 	ns.Add("Ik heb de krullen in het haar van mijn kat.")
	
	sentence, err := ns.Make()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("> %v", sentence)
}
