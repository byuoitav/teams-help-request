package main

import (
	"fmt"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"

	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/teams-help-request/goteamsnotification"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {

	var (
		logLevel   string
		hubAddress string
		webhookUrl string
		smeeUrl    string
	)

	// parse flags
	pflag.StringVarP(&logLevel, "log-level", "L", "info", "Level at which the logger operates")
	pflag.StringVarP(&hubAddress, "hub-address", "", "", "Address of the event hub")
	pflag.StringVarP(&webhookUrl, "webhook-url", "", "", "URL of the webhook to send to")
	pflag.StringVarP(&smeeUrl, "monitoring-url", "", "", "URL of the AV Monitoring Service")

	pflag.Parse()

	_, log := logger(logLevel)
	defer log.Sync()

	rm := goteamsnotification.RequestManager{
		Log:           log,
		MonitoringURL: smeeUrl,
		WebhookURL:    webhookUrl,
	}

	//Connecting to event hub
	log.Info("Starting event hub messenger")
	eventMessenger, nerr := messenger.BuildMessenger(hubAddress, base.Messenger, 5000)
	if nerr != nil {
		log.Fatal("failed to build event hub messenger", zap.Error(nerr))
	}

	//Subscribe to the event hub
	log.Info("Listening for room events")

	eventMessenger.SubscribeToRooms("*")

	for {
		event := eventMessenger.ReceiveEvent()
		if checkEvent(event) {
			log.Debug(fmt.Sprintf("this is a help request: %s", event.Key))

			//Send Message Via Teams
			rm.SendTheMessage(event.GeneratingSystem)
		}
	}
}

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
