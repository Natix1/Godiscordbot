package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	gatewayURL    = "wss://gateway.discord.gg/?v=10&encoding=json"
	intentsNumber = 33280
	botPrefix     = "!"
)

var (
	Authenticated  bool = false
	SequenceNumber int
	SessionId      string
	ResumeURL      string
	Bot            DiscordBot
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("Failed to obtain TOKEN from environment variables")
	} else {
		log.Printf("Token read from environment succesfully\n")
	}

	Bot.Token = token
}

func main() {
	conn, err := connect()
	if err != nil {
		log.Fatal("Dial failed:", err)
	}

	defer conn.Close()

	var helloEvent HelloEvent
	err = conn.ReadJSON(&helloEvent)

	if err != nil {
		log.Fatal(err)
	}

	if helloEvent.Opcode != 10 {
		log.Fatalln("Invalid discord response opcode; \nExpected 10;\nGot", helloEvent.Opcode)
	} else {
		log.Printf("Got opcode 10 (Hello) from discord\n")
	}

	go heartbeatRunner(conn, time.Duration(helloEvent.Data.HeartbeatIntervalMs)*time.Millisecond)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}

		var payload OpcodeBase
		err = json.Unmarshal(msg, &payload)
		if err != nil {
			log.Fatalf("%s\n", err.Error())
		}

		switch payload.Opcode {
		case 1: // Heartbeat
			log.Printf("Heartbeat requested from discord")
			sendHeartbeat(conn)

		case 11: // Heartbeat ACK
			log.Printf("Heartbeat acknowledged\n")

		case 0: // Event
			var event Event

			err = json.Unmarshal(msg, &event)
			if err != nil {
				log.Fatal(err)
			}

			SequenceNumber = event.SequenceNumber

			switch event.Type {
			case "READY": // Ready
				var readyEvent ReadyEvent

				err = json.Unmarshal(msg, &readyEvent)
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("Gateway ready, authenticated as %s with user ID being %s\n", readyEvent.Data.User.Username, readyEvent.Data.User.Id)
				Bot.User = readyEvent.Data.User
				Authenticated = true

			case "MESSAGE_CREATE":
				var message MessageEvent

				err = json.Unmarshal(msg, &message)
				if err != nil {
					log.Fatal(err)
				}

				go onMessage(&message)
			}

		default:
			log.Printf("Received opcode %d from discord, which was not handled by anything!", payload.Opcode)
		}

		if !Authenticated && payload.Opcode == 11 { // Attempt auth
			identify(conn)
		}

		log.Printf("Iteration done, sequence: %d\n", SequenceNumber)
	}
}
