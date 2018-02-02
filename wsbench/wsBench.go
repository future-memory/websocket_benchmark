package wsbench

import (
	_ "fmt"
	"github.com/gorilla/websocket"
	"log"
	"bytes"
	"net/url"
	_ "os"
	_ "os/signal"
	"sync"
	"sync/atomic"
	"time"
)

func WsBench(scheme string, address string, path string, sockets int, interval int, message string, duration int, connectionTimeout int, readonly int, auth string) {
	u := url.URL{Scheme: scheme, Host: address, Path: path}
	log.Println("%s", message);
	log.Printf("connecting to %s", u.String())
	start := time.Now()
	counter := 0
	var readCounter uint64 = 0
	var writeCounter uint64 = 0
	var compareError uint64 = 0
	var readError uint64 = 0
	var writeError uint64 = 0
	var connectionError uint64 = 0
	var writeBytes uint64 = 0
	var readBytes uint64 = 0
	var durr time.Duration
	var wg sync.WaitGroup

	authMsg := []byte(auth);

	for {
		counter++
		wg.Add(1)
		go func() {
			var dialer = websocket.Dialer{
				HandshakeTimeout: time.Duration(connectionTimeout) * time.Second,
			}
			co, _, err := dialer.Dial(u.String(), nil)

			if err != nil {
				log.Println("dial:", err)
				connectionError++
			} else {

				if(bytes.Count(authMsg, nil)>1){
					co.WriteMessage(websocket.TextMessage, authMsg)
				}

				defer co.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				for {
					if time.Since(start) > (time.Duration(duration) * time.Second) {
						break
					}
					writeTime := time.Now()
					if(readonly<1){
						err = co.WriteMessage(websocket.TextMessage, []byte(message))
						if err != nil {
							log.Println("write:", err)
							atomic.AddUint64(&writeError, 1)
						} 
					}

					atomic.AddUint64(&writeBytes, uint64(len([]byte(message))))
					atomic.AddUint64(&writeCounter, 1)
					_, readMessage, err := co.ReadMessage()
					if err != nil {
						log.Println("read:", err)
						readError++
					} else if readonly<1 && string(readMessage) != message {
						atomic.AddUint64(&compareError, 1)
						atomic.AddUint64(&readCounter, 1)
					} else {
						atomic.AddUint64(&readBytes, uint64(len([]byte(readMessage))))
						atomic.AddUint64(&readCounter, 1)
					}

					dur := time.Since(writeTime)
					log.Println(dur)
					durr += dur
					time.Sleep(time.Duration(interval) * time.Second)
				}
			}
			defer wg.Done()
		}()
		if counter >= sockets {
			break
		}
	}
	wg.Wait()
	readCounterF := atomic.LoadUint64(&readCounter)
	writeCounterF := atomic.LoadUint64(&writeCounter)
	writeBytesF := atomic.LoadUint64(&writeBytes)
	readBytesF := atomic.LoadUint64(&readBytes)
	connectionErrorF := atomic.LoadUint64(&connectionError)
	writeErrorF := atomic.LoadUint64(&writeError)
	readErrorF := atomic.LoadUint64(&readError)
	compareErrorF := atomic.LoadUint64(&compareError)
	var averageRtt time.Duration
	if readCounterF == 0 {
		averageRtt = time.Duration(0)
	} else {
		averageRtt = durr / time.Duration(readCounterF)
	}
	log.Println("Total Sent:", writeCounterF, ", Total Received:", readCounterF, ", Bytes Sent", writeBytesF, ", Bytes Received:", readBytesF, ", Average RTT:", averageRtt, ", Connection Error:", connectionErrorF, ", Write Error:", writeErrorF, ", Read Error:", readErrorF, ", Message Mismatch:", compareErrorF)
}
