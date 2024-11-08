package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// https://github.com/bencouture/denon-rest-api/blob/master/protocol.pdf
func main() {
	mux := http.NewServeMux()
	mux.Handle("/dolbyMovie", &DolbyHandler{})
	mux.Handle("/stereo", &StereoHandler{})
	mux.Handle("/direct", &DirectHandler{})
	mux.HandleFunc("/dolby", dolbyHandler)
	mux.HandleFunc("/music", musicHandler)
	mux.HandleFunc("/game", gameHandler)
	mux.HandleFunc("/tv", tvHandler)
	mux.HandleFunc("/roku/{command}",rokuHandler)
	mux.HandleFunc("/network", networkHandler)
	mux.HandleFunc("/cursor/{cursor}", cursorHandler)
	mux.HandleFunc("/menu/{state}", menuHandler)
	mux.HandleFunc("/input/{state}", inputHandler)
	mux.HandleFunc("/vol/{setting}", volHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.ListenAndServe(":8080", mux)

}

type DolbyHandler struct{}

func (d *DolbyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dolbyMovie()
}

type StereoHandler struct{}

func (d *StereoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stereo()
}

type DirectHandler struct{}

func (d *DirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	direct()
}

func rokuHandler(w http.ResponseWriter, r *http.Request) {
	cmd := strings.ToLower(r.PathValue("command"))
	sendCMD := ""
	switch cmd{
	case "poweron":
		sendCMD = "keypress/PowerOn"
	case "power":
		sendCMD = "keypress/Power"
	case "home":
		sendCMD = "keypress/Home"
	case "back":
		sendCMD = "keypress/Back"
	case "avr":
		sendCMD = "keypress/InputHdmi3"
	case "shield":
		sendCMD = "keypress/InputHdmi2"
	case "switch":
		sendCMD = "keypress/InputHdmi1"
	case "volumeup":
		sendCMD = "keypress/VolumeUp"
	case "volumedown":
		sendCMD = "keypress/VolumeDown"
	case "up":
		sendCMD = "keypress/Up"
	case "left":
		sendCMD = "keypress/Left"
	case "ok":
		sendCMD = "keypress/Select"
	case "right":
		sendCMD = "keypress/Right"
	case "down":
		sendCMD = "keypress/Down"
	default:
		return
	}

	url := "http://192.168.67.149:8060/" + sendCMD //TODO: Magic Number
	resp, err := http.Post(url,"application/json",bytes.NewBuffer(nil))
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		fmt.Println(msg)
		fmt.Fprintf(w,msg)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Response code:", resp.Status)
	fmt.Fprintf(w,"Response code: %s", resp.Status)
	}
func inputHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	state := r.PathValue("state")
	switch strings.ToLower(state) {
	case "on":
		sendCommand([]byte("MNSRC ON\r"), con)
		fmt.Fprintf(w, "<button class='darkGreen' hx-get='input/off' hx-swap='outerHTML'>Input</button>")
	case "off":
		sendCommand([]byte("MNSRC OFF\r"), con)
		fmt.Fprintf(w, "<button hx-get='input/on' hx-swap='outerHTML'>Input</button>")

	default:
		fmt.Fprintf(con, "Unknown menu Command: %s", state)
	}
}
func menuHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	state := r.PathValue("state")
	switch strings.ToLower(state) {
	case "on":
		sendCommand([]byte("MNMEN ON\r"), con)
		fmt.Fprintf(w, "<button class='darkGreen' hx-get='menu/off' hx-swap='outerHTML'>Menu</button>")
	case "off":
		sendCommand([]byte("MNMEN OFF\r"), con)
		fmt.Fprintf(w, "<button hx-get='menu/on' hx-swap='outerHTML'>Menu</button>")

	default:
		fmt.Fprintf(con, "Unknown menu Command: %s", state)
	}
}
func cursorHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	cursor := r.PathValue("cursor")
	switch strings.ToLower(cursor) {
	case "enter":
		sendCommand([]byte("MNENT\r"), con)
		fmt.Fprint(w, "MNENT")
	case "right":
		sendCommand([]byte("MNCRT\r"), con)
		fmt.Fprint(w, "MNCTT")
	case "left":
		sendCommand([]byte("MNCLT\r"), con)
		fmt.Fprint(w, "MNCLT")
	case "up":
		sendCommand([]byte("MNCUP\r"), con)
		fmt.Fprint(w, "MNUP")
	case "down":
		sendCommand([]byte("MNCDN\r"), con)
		fmt.Fprint(w, "MNDN")
	case "return":
		sendCommand([]byte("MNRTN\r"), con)
		fmt.Fprint(w, "MNRTN")
	default:
		fmt.Fprintf(w, "Invalid Command: %s", cursor)
	}
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	sendCommand([]byte("SITV\r"), con)
	sendCommand([]byte("MSDOLBY DIGITAL\r"), con)
	w.Write([]byte("Music Mode Activated"))
}

func dolbyHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	sendCommand([]byte("MSDOLBY DIGITAL\r"), con)
	w.Write([]byte("Dolby Digital Activated"))
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	r.SetPathValue("command","poweron")
	rokuHandler(w, r)
	time.Sleep(5 * time.Second)
	r.SetPathValue("command","avr")
	rokuHandler(w,r)
	time.Sleep(5 * time.Second)
	w.Write([]byte("GameMode Activated"))
	sendCommand([]byte("SIGAME\r"), con)
	sendCommand([]byte("MSMCH STEREO\r"), con)
}
func networkHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	r.SetPathValue("command","poweron")
	rokuHandler(w, r)
	time.Sleep(5 * time.Second)
	r.SetPathValue("command","avr")
	rokuHandler(w, r)
	time.Sleep(5 * time.Second)
	sendCommand([]byte("SINET\r"), con)
	sendCommand([]byte("MSDIRECT\r"), con)
}
func tvHandler(w http.ResponseWriter, r *http.Request) {
	con := avrConnect()
	defer con.Close()
	sendCommand([]byte("SITV\r"), con)
	sendCommand([]byte("MSMCH STEREO\r"), con)
	w.Write([]byte("TV Activated"))
}

func volHandler(w http.ResponseWriter, r *http.Request) {
	param := r.PathValue("setting")
	if strings.Contains(param, "+") {
		param = strings.ReplaceAll(param, "+", "")
		adjustment, err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse to int: ", param)
			w.Write([]byte(msg))
			return
		}
		volUp(adjustment)
		return
	} else if strings.Contains(param, "-") {
		param = strings.ReplaceAll(param, "-", "")
		adjustment, err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse to int: ", param)
			w.Write([]byte(msg))
			return
		}
		volDown(adjustment)
	} else {
		adjustment, err := strconv.Atoi(param)
		if err != nil {
			msg := fmt.Sprintf("Unable to parse int: ", param)
			w.Write([]byte(msg))
			return
		}
		volSet(adjustment)
	}
}

func avrConnect() net.Conn {
	const AVR_ADDR = "192.168.67.246:23"
	con, err := net.Dial("tcp", AVR_ADDR)
	if err != nil {
		log.Fatalf("Cannot connect to %s : %s\n", AVR_ADDR, err)
	}
	fmt.Println("Connected to AVR")
	return con
}

func volUp(x int) {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Increasing Volume:", x)
	data := []byte(fmt.Sprintf("MVUP\r"))
	fmt.Println(data)
	for i := 0; i < x*2; i++ {
		sendCommand(data, con)
	}
}
func volDown(x int) {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Decreasing Volume:", x)
	data := []byte(fmt.Sprintf("MVDOWN\r"))
	fmt.Println(data)
	for i := 0; i < x*2; i++ {
		sendCommand(data, con)
	}
}
func volSet(x int) {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Setting Volume to", x)
	data := []byte(fmt.Sprintf("MV%02d\r", x))
	fmt.Println(data)
	sendCommand(data, con)
}

func direct() {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Switching to Direct")
	data := []byte("MSDIRECT\r")
	sendCommand(data, con)
}

func stereo() {
	con := avrConnect()
	defer con.Close()

	fmt.Println("Switching to Stero")
	data := []byte("MSSTEREO\r")
	sendCommand(data, con)

	fmt.Println("Turning off Cinema EQ")
	data = []byte("PSCINEMA EQ.OFF\r")
	sendCommand(data, con)
}

func dolbyMovie() {
	con := avrConnect()
	defer con.Close()
	fmt.Println("Switching to TV Input")
	data := []byte("SITV\r")
	sendCommand(data, con)

	fmt.Println("Swithcing to Dolby Audio")
	data = []byte("MSDOLBY DIGITAL\r")
	sendCommand(data, con)

	fmt.Println("Switching to Cinema EQ")
	data = []byte("PSCINEMA EQ.ON\r")
	sendCommand(data, con)

	fmt.Println("Deactivating Loudness Management")
	data = []byte("PSLOM OFF\r")
	sendCommand(data, con)

	fmt.Println("Turning off Tone Control")
	data = []byte("PSTONE CTRL ON\r")
	sendCommand(data, con)

	fmt.Println("Set BASS +2")
	data = []byte("PSBAS 52\r")
	sendCommand(data, con)

	fmt.Println("Set TREB -2")
	data = []byte("PSTRE 48\r")
	sendCommand(data, con)
}

func sendCommand(data []byte, con net.Conn) {
	fmt.Printf("[TX]: %s\n", data)
	_, err := con.Write(data)
	if err != nil {
		log.Fatalf("[ERR] :: %s\n", err)
	}
	time.Sleep(20 * time.Millisecond)
}
