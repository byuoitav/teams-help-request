package goteamsnotification

import (
	"fmt"
	"log"
	"strings"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
)

func SendTheMessage(data string, webhook string, smee string) error {

	tokens := strings.Split(data, "-")
	building := tokens[0]
	room := tokens[1]
	device := tokens[2]

	// init client
	mstClient := goteamsnotify.NewClient()

	// setup webhook url
	webhookUrl := webhook

	// destination for OpenUri action
	smeeURL := fmt.Sprintf(smee+"%s-%s", building, room)
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
