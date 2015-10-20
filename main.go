package main

import (
	"log"
	"os"
	"strings"
	"./nonsentence"
	irc "github.com/fluffle/goirc/client"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %v nickname server channel", os.Args[0])
	}
	
	var nickname = os.Args[1]
	var server = os.Args[2]
	var channel = os.Args[3]
	
	log.Printf("Running Pipo");
	
	ns, err := nonsentence.New("nonsentence.db")
	if err != nil {
		log.Fatal(err)
	}
	defer ns.Close()
	
	quit := make(chan bool)
	
	client := irc.SimpleClient(nickname)
	
	client.HandleFunc(irc.CONNECTED, func (conn *irc.Conn, line *irc.Line) {
		conn.Join(channel)
	})
	
	client.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})
	
	client.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		log.Printf("Message to %v: %v", line.Target(), line.Text())
		// If a channel message is received, store it
		if line.Target() == channel {
			// Ignore first word if it ends with a ':'
			var words = strings.Fields(line.Text())
			if (len(words) > 0) && strings.HasSuffix(words[0], ":") {
				// If the message was directed at me, reply
				if words[0] == nickname + ":" {
					saySomething(client, channel, ns)
				}
				words = words[1:]
			}
			if err := ns.Add(strings.Join(words, " ")); err != nil {
				log.Printf("Error while adding sentence: %v", err)
			}
		} else if !strings.HasPrefix(line.Target(), "#") {
			// If a private message is received, say something
			saySomething(client, channel, ns)
		}
	})
	
	log.Printf("Connecting...")
	if err := client.ConnectTo(server); err != nil {
		log.Fatal(err)
	}
	defer client.Quit("Terminating")
	log.Printf("Connected!")
	
	<-quit
}

func saySomething(conn *irc.Conn, channel string, ns *nonsentence.Nonsentence) {
	sentence, err := ns.Make()
	if err != nil {
		log.Printf("Error while making sentence: %v", err)
	} else {
		conn.Privmsg(channel, sentence)
	}
}
