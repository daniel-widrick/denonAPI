package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

var con net.Conn

// https://github.com/bencouture/denon-rest-api/blob/master/protocol.pdf
func main(){
	mux := http.NewServeMux()
	mux.Handle("/dolbyMovie",&DolbyHandler{})
	http.ListenAndServe(":8080",mux)

}

type DolbyHandler struct{}
func(d *DolbyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const AVR_ADDR = "192.168.67.246:23"
	con, err := net.Dial("tcp",AVR_ADDR)
	if err != nil {
		log.Fatalf("Cannot connect to %s : %s\n",AVR_ADDR,err)
	}
	defer con.Close()
	fmt.Println("Connected to AVR")
	dolbyMovie(con)
}

func dolbyMovie(con net.Conn){
	fmt.Println("Switching to TV Input")
	data := []byte("SITV\r")
	sendCommand(data,con)

	fmt.Println("Swithcing to Dolby Audio")
	data = []byte("MSDOLBY DIGITAL\r")
	sendCommand(data,con)

	fmt.Println("Switching to Cinema EQ")
	data = []byte("PSCINEMA EQ.ON\r")
	sendCommand(data,con)

	fmt.Println("Deactivating Loudness Management")
	data = []byte("PSLOM OFF\r")
	sendCommand(data,con)

	fmt.Println("Turning off Tone Control")
	data = []byte("PSTONE CTRL ON\r")
	sendCommand(data,con)

	fmt.Println("Set BASS +2")
	data = []byte("PSBAS 52\r")
	sendCommand(data,con)

	fmt.Println("Set TREB -2")
	data = []byte("PSTRE 48\r")
	sendCommand(data,con)
}

func sendCommand(data []byte, con net.Conn){
	fmt.Printf("[TX]: %s\n",data)
	_, err := con.Write(data)
	if err != nil {
		log.Fatalf("[ERR] :: %s\n",err)
	}
	time.Sleep(2*time.Second)
}
