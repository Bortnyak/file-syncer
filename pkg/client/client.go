package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Bortnyak/file-syncer/pkg/storage"
	"github.com/r3labs/sse"
)

type UpdateEventPayload struct {
	Event string `json:"event"`
	Info  string `json:"info"`
}

func StartClient(ctx context.Context) error {
	err := listenToUpdates(ctx)
	if err != nil {
		return err
	}

	pullErr := pull()
	if pullErr != nil {
		return pullErr
	}

	return nil
}

func SendUpdate(payload *UpdateEventPayload) {
	client := http.Client{Timeout: 5 * time.Second}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: move server to env
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8085/sync", &buf)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: remove basic auth
	req.SetBasicAuth("admin", "admin123")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s\n", string(resBody))
}

func listenToUpdates(ctx context.Context) error {
	events := make(chan *sse.Event)
	client := getSSEClient()
	client.SubscribeChan("message", events)

	for {
		var message *sse.Event

		select {
		case message = <-events:
		case <-ctx.Done():
			return nil
		}

		fmt.Println("message from sse ")
		// fmt.Println("event: ", string(message.Event))
		dataStr := string(message.Data)
		fmt.Println("data: ", dataStr)

		if strings.Contains(dataStr, "Create") {
			path := strings.Split(dataStr, "/")
			fileName := path[len(path)-1]
			log.Println("File name to download from storage = ", fileName)

			storage.DownloadFile(fileName)
		}
	}
}

func getSSEClient() *sse.Client {
	// TODO: move server to env
	client := sse.NewClient("http://localhost:8085/stream")

	// TODO: remove basic auth
	username := "admin"
	password := "admin123"
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	client.Headers = map[string]string{
		"Authorization": "Basic " + auth,
	}

	return client
}

func pull() error {
	err := storage.GetList()
	if err != nil {
		return err
	}

	return nil
}
