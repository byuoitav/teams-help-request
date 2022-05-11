package main

import (
	"fmt"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/help-request-teams-notifier/notifier"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

// EventTags []string `json:"event-tags"` -> must contain alert
// Key string `json:"key"` -> must be help-request
// Value string `json:"value"` -> confirm ???

func main() {

	var (
		logLevel   string
		hubAddress string
		apiAddress string
		dbAddress  string
		roomList   []string
	)

	pflag.StringVarP(&logLevel, "log-level", "L", "info", "Level at which the logger operates")
	pflag.StringVarP(&roomList[0], "device-id", "", "", "Device id as found in couch")
	pflag.StringVarP(&hubAddress, "hub-address", "", "", "Address of the event hub")
	pflag.StringVarP(&apiAddress, "av-api", "", "", "Address of the av-api")
	pflag.StringVarP(&dbAddress, "db-address", "", "", "Address of the room database")
	pflag.Parse()

	_, log := logger(logLevel)
	defer log.Sync()

	//Connecting to event hub
	log.Info("Starting event hub message")
	eventMessenger, nerr := messenger.BuildMessenger(hubAddress, base.Messenger, 5000)
	if nerr != nil {
		log.Fatal("failed to build event hub messenger", zap.Error(nerr))
	}

	//Subscribe to the event hub
	log.Info("Listening for room events")
	for _, room := range roomList {
		eventMessenger.SubscribeToRooms(room)
	}

	notification := &notifier.Notification{
		Log:      log,
		Room:     "",
		Building: "",
		Device:   "",
		Class:    "",
	}

	for {
		event := eventMessenger.ReceiveEvent()
		if checkEvent(event) {
			log.Debug(fmt.Sprintf("this is a help request: %s", event.Key))

			//Handle Help Request
			//Send Message Via Teams
			notification.HandleEvent(event)
		}
	}
}

//Helper Functions

func checkEvent(event events.Event) bool {
	return event.Key == "help-request" && event.Value == "confirm" && contains(event.EventTags, "alert")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
