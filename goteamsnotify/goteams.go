package goteamsnotify

import (
	"fmt"
	"log"
	"strings"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
)

func main(data string) {
	_ = sendTheMessage(data)
}

func sendTheMessage(data string) error {

	tokens := strings.Split(data, "-")
	building := tokens[0]
	room := tokens[1]
	device := tokens[2]

	// init client
	mstClient := goteamsnotify.NewClient()

	// setup webhook url
	webhookUrl := "https://byu.webhook.office.com/webhookb2/5d065f4b-f069-4811-8f6a-4a26a12fe748@c6fc6e9b-51fb-48a8-b779-9ee564b40413/IncomingWebhook/e0f74f40e52e4360b5b9bd6e67945758/10d75264-373d-4eef-a190-d1c40162d969"

	// destination for OpenUri action
	smeeURL := fmt.Sprintf("https://newsmee.av.byu.edu/rooms/%s-%s", building, room)
	smeeURLDesc := "View Room in Monitoring"

	// setup message card
	msgCard := goteamsnotify.NewMessageCard()
	msgCard.Title = "Help Request Bot"
	msgCard.Summary = "Help Request Bot"

	info := goteamsnotify.NewMessageCardSection()
	info.ActivityTitle = fmt.Sprintf("Help Request for %s-%s", building, room)

	buildingFact := goteamsnotify.NewMessageCardSectionFact()
	buildingFact.Name = "Building"
	buildingFact.Value = building

	roomFact := goteamsnotify.NewMessageCardSectionFact()
	roomFact.Name = "Room"
	roomFact.Value = room

	deviceFact := goteamsnotify.NewMessageCardSectionFact()
	deviceFact.Name = "Device"
	deviceFact.Value = device

	info.AddFact(buildingFact, roomFact, deviceFact)

	msgCard.AddSection(info)

	// setup Action for message card
	smeeLink, err := goteamsnotify.NewMessageCardPotentialAction(
		goteamsnotify.PotentialActionOpenURIType,
		smeeURLDesc,
	)

	if err != nil {
		log.Fatal("error encountered when creating new action:", err)
	}

	smeeLink.MessageCardPotentialActionOpenURI.Targets =
		[]goteamsnotify.MessageCardPotentialActionOpenURITarget{
			{
				OS:  "default",
				URI: smeeURL,
			},
		}

	// add the Action to the message card
	if err := msgCard.AddPotentialAction(smeeLink); err != nil {
		log.Fatal("error encountered when adding action to message card:", err)
	}

	// send
	return mstClient.Send(webhookUrl, msgCard)
}
