package datachannel

import (
	"encoding/json"
	"log"
	"time"
)

const sleepTime = 5 * time.Second

type DataMessage struct {
	ID        int    `json:"id"`
	Data      string `json:"data"`
	Timestamp string `json:"timestamp"`
}
type Data struct {
	c chan string
}

func NewDataPump(c chan string) *Data {
	return &Data{
		c: c,
	}
}

func (d *Data) Start() {
	go func() {
		for i := 0; true; i++ {
			time.Sleep(sleepTime)
			currentTime := time.Now().Format(time.RFC3339)
			payload, err := json.Marshal(DataMessage{
				ID:        i,
				Data:      "ping",
				Timestamp: currentTime,
			})

			if err != nil {
				log.Println("error creating json object")

				continue
			}
			d.c <- string(payload)
		}
	}()
}
