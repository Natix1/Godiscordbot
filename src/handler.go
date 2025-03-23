package main

import (
	"fmt"
	"log"
)

func onMessage(Message *MessageEvent) {
	if Message.Data.Author.Id == Bot.Id {
		return
	}

	args, isCommand := CommandParser(Message.Data.Content)
	if !isCommand {
		log.Printf("%s said: %s", Message.Data.Author.Username, Message.Data.Content)
		return
	}

	switch args[0] {
	case "hello":
		content := fmt.Sprintf("<@%s>, echo: %s", Message.Data.Author.Id, Message.Data.Content)

		err := sendMessage(Message.Data.ChannelId, content)
		if err != nil {
			log.Printf("Failed sending message: %s", err.Error())
		}
	}
}
