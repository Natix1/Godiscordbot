package main

// op: 10 (https://discord.com/developers/docs/events/gateway#hello-event)
type HelloEvent struct {
	Opcode int `json:"op"`
	Data   struct {
		HeartbeatIntervalMs int `json:"heartbeat_interval"`
	} `json:"d"`
}

// general main goroutine refretch cycle response
type OpcodeBase struct {
	Opcode int `json:"op"`
}

// op: 1 (https://discord.com/developers/docs/events/gateway-events#heartbeat)
type Heartbeat struct {
	Opcode int `json:"op"`
	Data   any `json:"d"`
}

// op: 2 (https://discord.com/developers/docs/events/gateway-events#identify)
type Identify struct {
	Opcode int `json:"op"`
	Data   struct {
		Token      string `json:"token"`
		Properties struct {
			Os      string `json:"os"`
			Browser string `json:"browser"`
			Device  string `json:"device"`
		} `json:"properties"`
		Intents int `json:"intents"`
	} `json:"d"`
}

// op: 0 (https://discord.com/developers/docs/events/gateway-events#ready)
type Event struct {
	Type           string `json:"t"`
	Opcode         int    `json:"op"`
	SequenceNumber int    `json:"s"`
}

// op: 0 | READY
type ReadyEvent struct {
	Type           string `json:"t"`
	SequenceNumber int    `json:"s"`
	Opcode         int    `json:"op"`
	Data           struct {
		Version          int    `json:"v"`
		User             User   `json:"user"`
		SessionId        string `json:"session_id"`
		ResumeGatewayURL string `json:"resume_gateway_url"`
	} `json:"d"`
}

// Classes for discord events

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type DiscordBot struct {
	User
	Token string
}

type MessageEvent struct {
	Type           string `json:"t"`
	Opcode         int    `json:"op"`
	SequenceNumber int    `json:"s"`
	Data           struct {
		MessageId string `json:"id"`
		ChannelId string `json:"channel_id"`
		Author    User   `json:"author"`
		Content   string `json:"content"`
	} `json:"d"`
}
