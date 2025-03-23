package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	SendMessageURL = "https://discord.com/api/v10/channels/%s/messages"
)

func sendHeartbeat(conn *websocket.Conn) {
	data := Heartbeat{
		Opcode: 1,
		Data:   nil,
	}

	if SequenceNumber != 0 {
		data.Data = SequenceNumber
		log.Printf("Sending heartbeat to discord, with sequence number %d\n", SequenceNumber)
	} else {
		log.Printf("Sending heartbeat to discord, with sequence number not set\n")
	}

	conn.WriteJSON(data)
}

func heartbeatRunner(conn *websocket.Conn, each time.Duration) {
	for {
		sendHeartbeat(conn)
		jitter := time.Duration(rand.Float64()*0.1*float64(each)) - (each / 20)
		sleepTime := each + jitter
		time.Sleep(sleepTime)
	}
}

func sendMessage(channelId string, content string) error {
	// Validate before hitting discord
	if len(content) > 2000 {
		return fmt.Errorf("Length over 2000 characters, canceling send to discord")
	}

	body := map[string]string{
		"content": content,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	fullUrl := fmt.Sprintf(SendMessageURL, channelId)

	req, err := http.NewRequest("POST", fullUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+Bot.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Failed to send message, status: %d", resp.StatusCode)
	}

	return nil
}

func connect() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(gatewayURL+gatewayParams, nil)
	if err != nil {
		return nil, err
	}

	var helloEvent HelloEvent
	err = conn.ReadJSON(&helloEvent)

	if helloEvent.Opcode != 10 {
		log.Fatalf("Invalid discord response opcode; Expected 10; Got %d\n", helloEvent.Opcode)
	} else {
		log.Printf("Got opcode 10 (Hello) from discord\n")
	}

	go heartbeatRunner(conn, time.Duration(helloEvent.Data.HeartbeatIntervalMs)*time.Millisecond)

	return conn, nil
}

func resume(url string, sid string, seq int) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url+gatewayParams, nil)
	if err != nil {
		return nil, err
	}

	var helloEvent HelloEvent
	err = conn.ReadJSON(&helloEvent)

	if helloEvent.Opcode != 10 {
		log.Fatalf("Invalid discord response opcode; Expected 10; Got %d\n", helloEvent.Opcode)
	} else {
		log.Printf("Got opcode 10 (Hello) from discord\n")
	}

	data := ResumeEvent{
		Opcode: 6,
		Data: struct {
			Token          string "json:\"token\""
			Sessionid      string "json:\"session_id\""
			SequenceNumber int    "json:\"seq\""
		}{
			Token:          Bot.Token,
			Sessionid:      sid,
			SequenceNumber: seq,
		},
	}

	conn.WriteJSON(data)
	go heartbeatRunner(conn, time.Duration(helloEvent.Data.HeartbeatIntervalMs)*time.Millisecond)

	return conn, nil
}

func identify(conn *websocket.Conn) {
	identify := Identify{
		Opcode: 2,
		Data: struct {
			Token      string "json:\"token\""
			Properties struct {
				Os      string "json:\"os\""
				Browser string "json:\"browser\""
				Device  string "json:\"device\""
			} "json:\"properties\""
			Intents int "json:\"intents\""
		}{
			Token: Bot.Token,
			Properties: struct {
				Os      string "json:\"os\""
				Browser string "json:\"browser\""
				Device  string "json:\"device\""
			}{
				Os:      "linux",
				Browser: "golang",
				Device:  "WSL2",
			},
			Intents: intentsNumber,
		},
	}

	conn.WriteJSON(identify)
}

// Returns an array of strings which is the command and its arguments. The first argument is always the command WITHOUT the prefix, LOWERED
func CommandParser(message string) ([]string, bool) {
	sections := strings.Split(message, " ")
	sections[0] = strings.ToLower(sections[0])

	if string([]byte(sections[0])[0]) != botPrefix {
		return []string{""}, false
	}

	sections[0] = strings.TrimPrefix(sections[0], botPrefix)

	return sections, true
}
