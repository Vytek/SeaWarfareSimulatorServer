package main

import (
	"net"
	"strings"

	log "github.com/gookit/slog"
	"github.com/lrita/cmap"
	"github.com/obsilp/rmnp"
	"github.com/rs/xid"
	"github.com/vmihailenco/msgpack/v5"
	"gitlab.com/rwxrob/uniq"
	"github.com/common-nighthawk/go-figure"
)

var guid xid.ID
var n cmap.Map[string, *rmnp.Connection]

type Messages struct {
	OpCode  byte
	Message string
}

// Constants
const ServiceName = "Sea Warfare Simulator - SWS"
const ServiceVersion = "v0.1 Alfa"

func main() {
	//Unique ID
	guid = xid.New()
	log.Infof("Server uniqueID: %s\n", guid.String())

	server := rmnp.NewServer(":10001") //TODO: Add ini config for port and others

	server.ClientConnect = clientConnect
	server.ClientDisconnect = clientDisconnect
	server.ClientTimeout = clientTimeout
	server.ClientValidation = validateClient
	server.PacketHandler = handleServerPacket

	server.Start()
	log.Infof("Server started")
	myFigure := figure.NewFigure(ServiceName+" "+ServiceVersion, "", true)
	myFigure.Print()
	log.Infof("Service: %s Version: %s", ServiceName, ServiceVersion)

	select {}
}

func clientConnect(conn *rmnp.Connection, data []byte) {
	log.Infof("Client connection with: %s\n", data)
	//https://stackoverflow.com/questions/55959990/golang-print-string-as-an-array-of-bytes

	UniqueID := uniq.Hex(18)
	//Add new client connected
	n.Store(UniqueID+":"+conn.Addr.String(), conn)
	b, err := msgpack.Marshal(&Messages{OpCode: 1, Message: "cid:" + UniqueID + ":" + conn.Addr.String()})
	if err != nil {
		log.Errorf("Can't create MessagePack OpCode 1 Message")
	} else {
		conn.SendReliableOrdered(b)
		log.Infof("Send OpCode 1 Message to Client")
	}
}

func clientDisconnect(conn *rmnp.Connection, data []byte) {
	if len(data) == 0 {
		log.Infof("Client disconnect addr: %s\n", conn.Addr.String())
	} else {
		log.Infof("Client disconnect with: %s\n", data)
		//Parse Message received
		var MessageReceived Messages
		err := msgpack.Unmarshal(data, &MessageReceived)
		if err != nil {
			log.Errorf("Error clientDisconnect: %s\n", err)
			return
		}
		log.Infof(MessageReceived.Message)
		s := strings.Split(string(MessageReceived.Message), ":")
		if MessageReceived.OpCode == 2 {
			//Delete the client connected from cmap
			if s[0] == "dis" {
				n.Delete(s[1] + ":" + s[2] + ":" + s[3])
				log.Warnf("Client deleted: %s\n", s[1]+":"+s[2]+":"+s[3])
			} else {
				log.Errorf("Not dis in Message command")
			}
		} else {
			log.Errorf("Not Opcode 2 in Message")
		}
	}
}

func clientTimeout(conn *rmnp.Connection, data []byte) {
	if len(data) == 0 {
		log.Info("Client timeout")
	} else {
		log.Infof("Client timeout with data: %s\n", data)
	}
	//Delete the client Timeouted
	var ClientToDelete string
	n.Range(func(key string, value *rmnp.Connection) bool {
		k, v := key, value
		if v.Addr.String() == conn.Addr.String() {
			ClientToDelete = k
		}
		return true
	})
	n.Delete(ClientToDelete)
	log.Warnf("Client deleted: %s\n", ClientToDelete)
}

func validateClient(addr *net.UDPAddr, data []byte) bool {
	//Parse Message received
	var MessageReceived Messages
	err := msgpack.Unmarshal(data, &MessageReceived)
	if err != nil {
		log.Errorf("Error validateClient: %s\n", err)
		return false
	}
	log.Infof(MessageReceived.Message)
	s := strings.Split(string(MessageReceived.Message), ":")
	if MessageReceived.OpCode == 0 {
		if s[0] == "lng" {
			log.Infof("OK, client validated!")
			return true
			//Check login and password using scrypt

		} else {
			log.Errorf("Not lng in Message command")
			return false
		}
	} else {
		log.Errorf("Not Opcode 1 in Message")
		return false
	}
}

func handleServerPacket(conn *rmnp.Connection, data []byte, channel rmnp.Channel) {
	str := string(data)
	log.Infof("'"+str+"'", "from", conn.Addr.String(), "on channel", channel)

	//Parse MessagePack
	if len(data) != 0 {
		//Parse Message received
		var MessageReceived Messages
		err := msgpack.Unmarshal(data, &MessageReceived)
		if err != nil {
			log.Infof("Error validateClient: %s\n", err)
			return
		}
		log.Infof(MessageReceived.Message)
		s := strings.Split(string(MessageReceived.Message), ":")
		if MessageReceived.OpCode == 3 {
			if s[0] == "ping" {
				log.Infof("OK, ping in Message command\n")
				conn.SendReliableOrdered(SendMessagePong())
			} else {
				log.Errorf("Not cid in Message command\n")
			}
		}
		//Others OpCodes
	}
}

func SendMessagePong() []byte {
	b, err := msgpack.Marshal(&Messages{OpCode: 3, Message: "pong"})
	if err != nil {
		panic(err)
	} else {
		return b
	}
}