package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Data struct {
	ID    string
	Type  string
	Value Value
}

type Value struct {
	Timestamp int64
	Value     interface{}
}

type DictionaryColection struct {
	Type  string     `json:"type"`
	Value Dictionary `json:"value"`
}

// Dictionary ...
type Dictionary struct {
	IDentifier string `json:"identifier"` // sc
	Name       string `json:"name"`       // Example Spacecraft
	Subsystems []struct {
		IDentifier   string `json:"identifier"` // prop
		Name         string `json:"name"`       // Propulsion
		Measurements []struct {
			IDentifier string `json:"identifier"` // prop.fuel
			Name       string `json:"name"`       // Fuel
			Type       string `json:"type"`       // float
			Units      string `json:"units"`      // kilograms
		} `json:"measurements"`
	} `json:"subsystems"`
}

type Client struct {
	ID int64
}

var (
	connList map[*websocket.Conn]Client
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func sendData(data Data) {
	fmt.Printf("ID: %s; Type: %s; Value: %v \n", data.ID, data.Type, data.Value.Value)
}

func NewData(id string, dataType string, value Value) Data {
	return Data{id, dataType, value}
}

func NewValue(timestamp int64, value interface{}) Value {
	return Value{timestamp, value}
}

func sender() {
	for {
		time.Sleep(time.Millisecond * time.Duration(2000))

		value := NewValue(makeTimestamp(), makeTimestamp())
		data := NewData("pwr.v", "data", value)

		fdata, _ := json.Marshal(data)
		fmt.Println(string(fdata))
	}
}

func session(conn *websocket.Conn) {

	defer func() {
		conn.Close()
		delete(connList, conn)
	}()

	for {
		var message string

		if err := websocket.Message.Receive(conn, &message); err != nil {
			log.Println("Error:", err.Error())
		}

		switch message {
		case "dictionary":

			var content []byte
			content, err := ioutil.ReadFile("./dictionary.json")
			if err != nil {
				panic(err)
			}

			fmt.Println(string(content))

			if err := websocket.Message.Send(conn, string(content)); err != nil {
				log.Println("Can't send message!")
				panic(err)
			}

		}

	}
}

func main() {

	connList := make(map[*websocket.Conn]Client)

	go sender()

	onConnected := func(conn *websocket.Conn) {

		connList[conn] = Client{makeTimestamp()}
		go session(conn)

	}

	http.Handle("/", websocket.Handler(onConnected))

	http.ListenAndServe(":8000", nil)

}
