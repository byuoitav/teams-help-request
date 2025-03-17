package goteamsnotification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type RequestManager struct {
	Log           *slog.Logger
	MonitoringURL string
	WebhookURL    string
}

// MessageCard is the structure for the Teams message
type MessageCard struct {
	ThemeColor string           `json:"themeColor"`
	Context    string           `json:"@context"`
	Sections   []MessageSection `json:"sections"`
	Summary    string           `json:"summary"`
	Type       string           `json:"@type"`
}

// MessageSection represents a section in the Teams message
type MessageSection struct {
	ActivityTitle    string     `json:"activityTitle"`
	ActivitySubtitle string     `json:"activitySubtitle"`
	Facts            []CardFact `json:"facts"`
	Markdown         bool       `json:"markdown"`
}

// CardFact represents a fact in a message section
type CardFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (rm *RequestManager) SendTheMessage(data string) error {
	tokens := strings.Split(data, "-")
	building := tokens[0]
	room := tokens[1]
	device := tokens[2]

	// destination for OpenUri action
	smeeURL := fmt.Sprintf(rm.MonitoringURL+"/rooms/%s-%s", building, room)

	// Create the message card payload
	msgCard := &MessageCard{
		ThemeColor: "0076D7",
		Context:    "http://schema.org/extensions",
		Summary:    "Help Request Bot",
		Type:       "MessageCard",
		Sections: []MessageSection{
			{
				ActivityTitle:    "Help Request Bot",
				ActivitySubtitle: fmt.Sprintf("Help Request for %s-%s", building, room),
				Facts: []CardFact{
					{
						Name:  "Building",
						Value: building,
					},
					{
						Name:  "Room",
						Value: room,
					},
					{
						Name:  "Device",
						Value: device,
					},
					{
						Name:  "Time Stamp",
						Value: time.Now().Format(time.RFC1123),
					},
				},
				Markdown: true,
			},
		},
	}

	// Convert the card to JSON
	jsonData, err := json.Marshal(msgCard)
	if err != nil {
		rm.Log.Error("error marshaling message card to JSON", "error", err)
		return err
	}

	// Debug log the JSON payload
	rm.Log.Debug("sending message to Teams webhook",
		"json_payload", string(jsonData),
		"webhook_url", rm.WebhookURL)

	// Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, rm.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		rm.Log.Error("error creating HTTP request", "error", err)
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Log the full HTTP request in debug mode
	if reqDump, err := httputil.DumpRequestOut(req, true); err == nil {
		rm.Log.Debug("outgoing HTTP request", "request", string(reqDump))
	}

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		rm.Log.Error("error sending request to Teams webhook", "error", err)
		return err
	}
	defer resp.Body.Close()

	// Log the full HTTP response in debug mode
	if respDump, err := httputil.DumpResponse(resp, true); err == nil {
		rm.Log.Debug("incoming HTTP response", "response", string(respDump))
	} else {
		// If we can't dump the whole response, at least try to log the body
		if respBody, err := io.ReadAll(resp.Body); err == nil {
			rm.Log.Debug("HTTP response body", "body", string(respBody))
		}
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		rm.Log.Error("received non-success status code from Teams webhook",
			"status_code", resp.StatusCode)
		return fmt.Errorf("received non-success status code %d from Teams webhook", resp.StatusCode)
	}

	rm.Log.Info("successfully sent Teams notification",
		"building", building,
		"room", room,
		"monitoring_url", smeeURL)

	return nil
}
