package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type msg struct {
	Num int
}

// DictionaryColection ...
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

type clientStatus struct {
	Subscriptions map[string]bool
}

// Data ...
type Data struct {
	ID    string    `json:"id"`   // pwr.v
	Type  string    `json:"type"` // data
	Value dataValue `json:"value"`
}

type dataValue struct {
	Timestamp int64       `json:"timestamp"` // 1474891025120
	Value     interface{} `json:"value"`     // 30.00046680219344
}

const dictionary = "dictionary.json"

var clientList map[*websocket.Conn]clientStatus
var systemStatus map[string]dataValue

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func makeData(id string, timestamp int64, value interface{}) (r Data) {
	r = Data{}

	r.ID = id
	r.Type = "data"
	r.Value.Timestamp = timestamp
	r.Value.Value = value

	return
}

func CloseAll() {
	for c := range clientList {
		c.Close()
	}
}

func ListenAndServe(port int, timerInterval int) {

	log.Println("Initializing Telemetry Hub")
	log.Println("websocket port", port)

	clientList = make(map[*websocket.Conn]clientStatus)
	systemStatus = make(map[string]dataValue)

	var aux dataValue
	aux.Timestamp = makeTimestamp()
	aux.Value = float64(1.0)
	systemStatus["pwr.v"] = aux

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(timerInterval))

			// remover
			var aux dataValue = systemStatus["pwr.v"]
			aux.Timestamp = makeTimestamp()
			x := aux.Value.(float64) + 0.1
			aux.Value = x
			systemStatus["pwr.v"] = aux

			fmt.Println("pwr.v = ", aux.Value.(float64))

			for id, status := range systemStatus {
				SendValue(id, status.Timestamp, status.Value)
			}
		}
	}()

	http.HandleFunc("/", wsHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func SendValue(id string, timestamp int64, value interface{}) {
	for conn, c := range clientList {
		if v, ok := c.Subscriptions[id]; ok && v {

			var p = makeData(id, timestamp, value)

			if err := conn.WriteJSON(p); err != nil {
				log.Println(err)
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//if r.Header.Get("Origin") != "http://"+r.Host {
	//	http.Error(w, "Origin not allowed", 403)
	//	return
	//}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	log.Println("Client connection")
	clientList[conn] = clientStatus{Subscriptions: make(map[string]bool)}

	log.Println("Client tot:", len(clientList))

	go telemetryWs(conn)
}

func telemetryWs(conn *websocket.Conn) {

	defer func() {
		conn.Close()
		delete(clientList, conn)
	}()

	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			return
		}

		msg := string(p)
		log.Printf("msg:[%s] type:%d\r\n", msg, messageType)
		msgArray := strings.Split(msg, " ")
		for k, v := range msgArray {
			log.Printf("%d\t[%s]\r\n", k, v)
		}

		switch msgArray[0] {
		case "teste":

			var m map[string]string
			m = make(map[string]string)
			m["teste1"] = "teste1"
			m["teste2"] = "teste2"

			if err = conn.WriteJSON(m); err != nil {
				log.Println(err)
			}

		case "dictionary":
			//var m map[string]string
			//m = make(map[string]string)
			//m["dictionary"] = readFile("./dictionary.json")

			/*
				type: "dictionary",
				value: dictionary
			*/
			var content []byte
			content, err = ioutil.ReadFile("./dictionary.json")

			if err != nil {
				panic(err)
			}

			doc := Dictionary{}
			err = json.Unmarshal(content, &doc)
			if err != nil {
				panic(err)
			}

			dc := DictionaryColection{}
			dc.Type = "dictionary"
			dc.Value = doc

			log.Println("Name:", dc.Value.Name)

			if err = conn.WriteJSON(dc); err != nil {
				log.Println(err)
			}
		case "subscribe":
			if len(msgArray) < 2 {
				log.Println("error: no subscribe parameter")
				//return
			}
			clientList[conn].Subscriptions[msgArray[1]] = true
			log.Println("client subscribe", msgArray[1])
		case "unsubscribe":
			if len(msgArray) < 2 {
				log.Println("error: no unsubscribe parameter")
				//return
			}
			clientList[conn].Subscriptions[msgArray[1]] = false
		case "history":

		case "list":
			var s []string
			for k, v := range clientList[conn].Subscriptions {
				if v {
					s = append(s, k)
				}
				if err = conn.WriteJSON(s); err != nil {
					log.Println(err)
				}
			}
		}

		//err = conn.WriteMessage(messageType, p)
		//if err != nil {
		//	return
		//}

		//if err = conn.WriteJSON(m); err != nil {
		//	fmt.Println(err)
		//}
	}
}

func readFile(fn string) string {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("read file:", fn, "error:", err)
		os.Exit(-1)
	}
	return string(content)
}
