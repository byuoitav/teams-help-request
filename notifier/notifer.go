package notifier

import (
	"fmt"
	"strings"

	"github.com/byuoitav/common/v2/events"
	"go.uber.org/zap"
)

/*
Help Request for ITB 1010

        Building
        ITB

        Room
        1010

        Device
        CP1

        Current Class
        None

*/

type Notification struct {
	Log      *zap.Logger
	Building string
	Room     string
	Device   string
	Class    string
}

func (n *Notification) HandleEvent(event events.Event) Notification {
	tokens := strings.Split(event.GeneratingSystem, "-")
	n.Building = tokens[0]
	n.Room = tokens[1]
	n.Device = tokens[2]
	n.Log.Debug(fmt.Sprintf("Help Request: %s-%s-%s", tokens[0], tokens[1], tokens[2]))

	return *n
}
