package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type Message struct {
	From    string `json:"From"`
	To      string `json:"To"`
	Payload []byte `json:"Payload"`
}

type Server struct {
	ln       net.Listener
	messages map[string][]Message
}

func NewServer() *Server {
	return &Server{
		messages: make(map[string][]Message),
	}
}

func (server *Server) recieve(w http.ResponseWriter, r *http.Request) {

}

func (server *Server) fetch(w http.ResponseWriter, r *http.Request) {

}

func main() {
	//ok
	// server := NewServer()

	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	contextBackground := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ping, err := redisClient.Ping(contextBackground).Result()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(ping)

	jsonString, err := json.Marshal(Message{
		From:    "Alice",
		To:      "Jason",
		Payload: []byte(`{"text": "JASON!! We should hack the pentagon!"}`),
	})

	// var msg Message

	if err != nil {
		fmt.Println("Failed to marshal:", err.Error())
	}
	// redisClient.Append(contextBackground, "Alice", "Hi")
	err = redisClient.Set(contextBackground, "Alice", jsonString, 0).Err()
	if err != nil {
		fmt.Println("Failed to set value in the redis instance: &s", err.Error())
		return
	}

	val, err := redisClient.Get(contextBackground, "Alice").Result()

	if err != nil {
		fmt.Println("failed to get value from get redis", err.Error())
		return
	}

	fmt.Println("marshalled:", string(val))

	var msg Message
	err1 := json.Unmarshal(jsonString, &msg)
	if err1 != nil {
		return
	}

	fmt.Println("unmarshalled:")
	fmt.Println("Message sender:", msg.From)
	fmt.Println("Message recipient:", msg.To)
	fmt.Println(string(msg.Payload))
	// fmt.Println(val.From)
	// fmt.Println(val.To)

	// var payloadData map[string]interface{}
	// err2 := json.Unmarshal(msg.Payload, &payloadData)
	// if err2 != nil {
	// 	return
	// }

	// fmt.Println(payloadData["text"])

	// fmt.Println("Starting server.go")
	// http.HandleFunc("/msg", server.recieve)
	// http.HandleFunc("/fetch", server.fetch)

	// http.ListenAndServe(":8080", nil)

}
