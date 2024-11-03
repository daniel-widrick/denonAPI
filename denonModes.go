package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// https://github.com/bencouture/denon-rest-api/blob/master/protocol.pdf
func main(){
	mux := http.NewServeMux()
	mux.Handle("/dolbyMovie",&DolbyHandler{})
	mux.Handle("/stereo",&StereoHandler{})
	mux.Handle("/direct",&DirectHandler{})
	mux.HandleFunc("/vol/{setting}",volHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w,r,"index.html")
	})
	http.ListenAndServe(":8080",mux)

}

type DolbyHandler struct{}
func(d *DolbyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dolbyMovie()
}

type StereoHandler struct{}
func(d *StereoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stereo()
}

type DirectHandler struct{}
func(d *DirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	direct()
}

func volHandler(w http.ResponseWriter, r *http.Request){
	param := r.PathValue("setting")
	if strings.Contains(param, "+"){
		param = strings.ReplaceAll(param,"+","")
		adjustment,err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse to int: ",param)
			w.Write([]byte(msg))
			return
		}
		volUp(adjustment)
		return
	} else if strings.Contains(param, "-"){
		param = strings.ReplaceAll(param,"-","")
		adjustment, err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse to int: ",param)
			w.Write([]byte(msg))
			return
		}
		volDown(adjustment)
	} else {
		adjustment,err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse int: ",param)
			w.Write([]byte(msg))
			return
		}
		volSet(adjustment)
	}
}

func avrConnect() net.Conn {
	const AVR_ADDR = "192.168.67.246:23"
	con, err := net.Dial("tcp",AVR_ADDR)
	if err != nil {
		log.Fatalf("Cannot connect to %s : %s\n",AVR_ADDR,err)
	}
	fmt.Println("Connected to AVR")
	return con
}

func volUp(x int){
	con := avrConnect()
	defer con.Close()

	fmt.Println("Increasing Volume:",x)
	data := []byte(fmt.Sprintf("MVUP\r"))
	fmt.Println(data)
	for i := 0; i < x*2; i++{
		sendCommand(data, con)
	}
}
func volDown(x int){
	con := avrConnect()
	defer con.Close()

	fmt.Println("Decreasing Volume:",x)
	data := []byte(fmt.Sprintf("MVDOWN\r"))
	fmt.Println(data)
	for i := 0; i < x*2; i++{
		sendCommand(data, con)
	}
}
func volSet(x int){
	con := avrConnect()
	defer con.Close()

	fmt.Println("Setting Volume to",x)
	data := []byte(fmt.Sprintf("MV%02d\r",x))
	fmt.Println(data)
	sendCommand(data,con)
}

func direct() {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Switching to Direct")
	data := []byte("MSDIRECT\r")
	sendCommand(data,con)
}

func stereo(){
	con := avrConnect()
	defer con.Close()

	fmt.Println("Switching to Stero")
	data := []byte("MSSTEREO\r")
	sendCommand(data,con)

	fmt.Println("Turning off Cinema EQ")
	data = []byte("PSCINEMA EQ.OFF\r")
	sendCommand(data,con)
}

func dolbyMovie(){
	con := avrConnect()
	defer con.Close()
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
	time.Sleep(20*time.Millisecond)
}
