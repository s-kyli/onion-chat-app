package client

import (
	"bufio"
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	From    string `json:"From"`
	To      string `json:"To"`
	Payload []byte `json:"Payload"`
}

type Contact struct {
	publicXKeyHex  string
	publicEdKeyHex string
	alias          string
	sharedSecret   []byte
}

type Client struct {
	listenAddr     string
	publicEdKeyHex string
	privateEdKey   ed25519.PrivateKey
	publicXKey     *ecdh.PublicKey
	publicXKeyHex  string
	privateXKey    *ecdh.PrivateKey   // ermm doesnnt look so secure, changge this later
	contacts       map[string]Contact // Key is the alias of contact
	serverUrl      string
}

func NewClient(listenAddr string) *Client {

	pubEdKey, privEdKey, _ := ed25519.GenerateKey(rand.Reader)
	pubEdKeyHex := hex.EncodeToString(pubEdKey)

	privXKey, _ := ecdh.X25519().GenerateKey(rand.Reader)
	pubXKey := privXKey.PublicKey()

	return &Client{
		listenAddr:     listenAddr,
		publicEdKeyHex: pubEdKeyHex,
		privateEdKey:   privEdKey,
		publicXKey:     pubXKey,
		publicXKeyHex:  hex.EncodeToString(pubXKey.Bytes()),
		privateXKey:    privXKey,
		contacts:       make(map[string]Contact), // key is alias of contact
		serverUrl:      "http://localhost:8080/send",
	}
}

func (c *Client) addContact(alias string, publicXKeyHex string, publicEdKeyHex string) {

	decodedPublicXKey, err := hex.DecodeString((publicXKeyHex))
	if err != nil {
		fmt.Println("Error decoding x25519 key:", err)
		return
	}

	pubXKey, err := ecdh.X25519().NewPublicKey(decodedPublicXKey)
	if err != nil {
		fmt.Println("Error converting decodedPublicXKey to ecdh.PublicKey", err)
		return
	}

	sharedSecret, err := c.privateXKey.ECDH(pubXKey)
	if err != nil {
		fmt.Println("Error creating shared secret", err)
		return
	}

	c.contacts[alias] = Contact{
		publicXKeyHex:  publicXKeyHex,
		publicEdKeyHex: publicEdKeyHex,
		alias:          alias,
		sharedSecret:   sharedSecret,
	}
}

func (c *Client) sendMessage(contact Contact, msgText string) {

	ciphertext, err := EncryptPayload(contact.sharedSecret, msgText)
	if err != nil {
		fmt.Println("sendMesesage failed, encryption failed:", err)
		return
	}

	jsonByte, err := MakeJsonByte(c.publicXKeyHex, contact.publicEdKeyHex, ciphertext)
	if err != nil {
		fmt.Println("makeJsonByte error:", err)
		return
	}

	// in the future, communication with the server will only be via Tor.
	response, err := http.Post(c.serverUrl, "application/json", bytes.NewBuffer(jsonByte))
	if err != nil {
		fmt.Println("HTTP POST failed:", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Server rejected message send. Status:", response.Status)
		return
	}

	fmt.Println("Message send to server successfully")

}

func (c *Client) processInput(text string) {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "/add-contact") {
		parts := strings.Split(text, " ")
		if len(parts) == 4 {
			c.addContact(parts[1], parts[2], parts[3])
		} else {
			fmt.Println("usage: /add-contact [alias (string)] [public X25519 key (hex string)] [public  ED25519 key (hex string)]")
		}
	} else if strings.HasPrefix(text, "/get-contacts") {
		for _, contact := range c.contacts {
			fmt.Printf("%s's contact info:\n", contact.alias)
			fmt.Println("Public X25519 key:", contact.publicXKeyHex)
			fmt.Println("Public ED25519 key:", contact.publicEdKeyHex)
			fmt.Println()
		}
	} else if strings.HasPrefix(text, "/chat") {
		parts := strings.SplitN(text, " ", 3)
		if len(parts) >= 3 {

			//just for now, adding encryption and communication to server.go later
			contact, exist := c.contacts[parts[1]]
			if exist {
				// fmt.Println("Chatted '", parts[2], "' to", contact.alias, "aka", contact.publicEdKeyHex)
				c.sendMessage(contact, parts[2])
			} else {
				fmt.Println(parts[1], "is not a valid contact")
			}

		}
	} else if strings.HasPrefix(text, "/disconnect") {
		parts := strings.Split(text, " ")
		if len(parts) == 2 {
			//tbd
		}
	}
}

func RunClient(args []string) {

	if len(args) == 0 {
		fmt.Println("usage: go run client_main.go [port]")
		return
	}

	//create test public key for testing with fake contact
	// privXKey, _ := ecdh.X25519().GenerateKey(rand.Reader)
	// fmt.Println("TEST PUBLIC X25519 KEY:")
	// fmt.Println(hex.EncodeToString(privXKey.PublicKey().Bytes()))

	// pubEdKey, _, _ := ed25519.GenerateKey(rand.Reader)
	// fmt.Println("TEST PUBLIC ED25519 KEY")
	// fmt.Println(hex.EncodeToString(pubEdKey))

	client := NewClient(":" + args[0])
	fmt.Println("YOUR PUBLIC X25519 KEY:")
	fmt.Println(hex.EncodeToString(client.publicXKey.Bytes()))

	fmt.Println("YOUR PUBLIC ED25519 KEY")
	fmt.Println(client.publicEdKeyHex)

	fmt.Println("\nCOMMANDS:")
	fmt.Println("/add-contact [alias] [public X25519 key] [public ED25519 key]")
	fmt.Println("/chat [contact alias] [message]")
	fmt.Println("/get-contacts")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		client.processInput(scanner.Text())
	}
}
