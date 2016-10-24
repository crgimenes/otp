package openTelemetryProvider

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
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

var port int
var timerInterval int
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

func main() {

	flag.IntVar(&port, "port", 8081, "websocket port, provides telemetry data to OpenMCT")
	flag.IntVar(&timerInterval, "timer", 1000, "main timer interval")
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	log.Println("Initializing Telemetry Hub")
	log.Println("websocket port", port)

	clientList = make(map[*websocket.Conn]clientStatus)
	systemStatus = make(map[string]dataValue)

	//var aux dataValue
	//aux.Timestamp = makeTimestamp()
	//aux.Value = float64(1.0)
	//systemStatus["pwr.v"] = aux

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		<-sc

		//TODO free resources ...

		// close all connections
		for c := range clientList {
			c.Close()
		}

		fmt.Print("\n")
		log.Println("Have a nice day!")
		os.Exit(0)
	}()

	/*
		//data
		type PwrV struct {
			ID    string `json:"id"`   // pwr.v
			Type  string `json:"type"` // data
			Value struct {
				Timestamp int64   `json:"timestamp"` // 1474891025120
				Value     interface{}  `json:"value"`     // 30.00046680219344
			} `json:"value"`
		}

		{"type":"data","id":"pwr.v","value":{"timestamp":1474891025120,"value":30.00046680219344}}
		{"type":"data","id":"pwr.v","value":{"timestamp":1475620104145,"value":30.012436195447858}}
		{"type":"data","id":"pwr.v","value":{"timestamp":1475620105154,"value":30.014189050863177}}
		{"type":"data","id":"pwr.v","value":{"timestamp":1475620106157,"value":30.02952627354623}}


		//history data
		type HpwrV struct {
			ID    string `json:"id"`   // pwr.v
			Type  string `json:"type"` // history
			Value []struct {
				Timestamp int64   `json:"timestamp"` // 1474891025120
				Value     interface{} `json:"value"`     // 30.086271694110074
			} `json:"value"`
		}

		//Example: Trying to access Action member of a statement myStatement.
		switch a := PwrV.Value.Value.(type) {
		case []string:
			//Action is a slice. Handle it accordingly.
		case string:
			//Action is a string. Handle it accordingly.
		default:
			//Some other datatype that can be returned by aws?
		}

		https://newfivefour.com/golang-interface-type-assertions-switch.html
		var anything interface{} = "string"
		switch v := anything.(type) {
		                    case string:
		                            fmt.Println(v)
		                    case int32, int64:
		                            fmt.Println(v)
		                    case SomeCustomType:
		                            fmt.Println(v)
		                    default:
		                            fmt.Println("unknown")
		            }

		http://attilaolah.eu/2013/11/29/json-decoding-in-go/
	*/

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(timerInterval))

			for id, status := range systemStatus {
				sendValue(id, status.Timestamp, status.Value)
			}
		}
	}()

	log.Println("Press ^C to terminate.")

	http.HandleFunc("/", wsHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func sendValue(id string, timestamp int64, value interface{}) {
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
		//fmt.Println("msg:", msg, messageType)
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
