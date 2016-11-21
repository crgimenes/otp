package openTelemetryProvider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

type handleFn func()

var clientList map[*websocket.Conn]clientStatus
var systemStatus map[string]dataValue
var subsystemHandleList map[string]handleFn
var taxonomyDictionary = Dictionary{}

//var lock = sync.RWMutex{}

func init() {
	log.Println("Initializing Telemetry Hub")

	clientList = make(map[*websocket.Conn]clientStatus)
	systemStatus = make(map[string]dataValue)
	subsystemHandleList = make(map[string]handleFn)
}

// MakeTimestamp creates an int64 with current timestamp in the format expected by OpenMCT
func MakeTimestamp() int64 {
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

// CloseAll close all websocket connections
func CloseAll() {
	for c := range clientList {
		c.Close()
	}
}

// SubsystemHandleFunc set a subsystem handle function
func SubsystemHandleFunc(subsystemIdentifier string, handleFunc handleFn) {
	subsystemHandleList[subsystemIdentifier] = handleFunc
}

// SetDataValue set a value to subsystem
func SetDataValue(identifier string, timeStamp int64, value interface{}) {
	dv := dataValue{}

	if timeStamp == -1 {
		timeStamp = MakeTimestamp()
	}

	dv.Timestamp = timeStamp
	dv.Value = value

	//lock.Lock()
	systemStatus[identifier] = dv
	//lock.Unlock()
}

// ListenAndServe websocket
func ListenAndServe(port int, timerInterval int) {

	log.Println("websocket port", port)

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(timerInterval))

			for subsystem := range subsystemHandleList {
				RunSubsystemHandleFunc(subsystem)
			}

			for id, status := range systemStatus {
				SendValue(id, status.Timestamp, status.Value)
			}

		}
	}()

	http.HandleFunc("/", wsHandler)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// SendValue send value to subsystem by identifier
func SendValue(id string, timestamp int64, value interface{}) {
	for conn, c := range clientList {
		//lock.Lock()
		//defer lock.Unlock()

		if v, ok := c.Subscriptions[id]; ok && v {

			var p = makeData(id, timestamp, value)

			if err := conn.WriteJSON(p); err != nil {
				log.Println(err)
			}
		}
	}
}

// RunSubsystemHandleFunc call subsystem handle function
func RunSubsystemHandleFunc(subsystemIdentifier string) {
	if f, ok := subsystemHandleList[subsystemIdentifier]; ok && f != nil {
		f()
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
		case "dictionary":
			dc := DictionaryColection{}
			dc.Type = "dictionary"
			dc.Value = taxonomyDictionary

			if err = conn.WriteJSON(dc); err != nil {
				log.Println(err)
			}
		case "subscribe":
			if len(msgArray) < 2 {
				log.Println("error: no subscribe parameter")
			}
			clientList[conn].Subscriptions[msgArray[1]] = true
			log.Println("client subscribe", msgArray[1])
		case "unsubscribe":
			if len(msgArray) < 2 {
				log.Println("error: no unsubscribe parameter")
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

// LoadTaxonomyDictionaryFromFile load dictionary from file
func LoadTaxonomyDictionaryFromFile(fn string) (err error) {
	var content []byte
	content, err = ioutil.ReadFile(fn)
	if err != nil {
		err = fmt.Errorf("LoadTaxonomyDictionaryFromFile error: %s", err)
		return
	}

	err = json.Unmarshal(content, &taxonomyDictionary)
	if err != nil {
		err = fmt.Errorf("LoadTaxonomyDictionaryFromFile error: %s", err)
		return
	}
	return
}
