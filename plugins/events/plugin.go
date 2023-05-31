package alpsevents

import (
	"fmt"
	"net/http"

	"git.sr.ht/~migadu/alps"
	"github.com/emersion/go-imap/client"
	"github.com/gorilla/websocket"
)

func formatUpdate(update client.Update) (string, error) {
	switch u := update.(type) {
	case *client.MailboxUpdate:
		return fmt.Sprintf("mailbox: %s", u.Mailbox.Name), nil
	case *client.MessageUpdate:
		return fmt.Sprintf("message: %d", u.Message.Uid), nil
	case *client.StatusUpdate:
		return fmt.Sprintf("status: %s", u.Status.Info), nil
	default:
		return "", fmt.Errorf("Uknown type")
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections
		return true
	},
}

func handleWebSocket(conn *websocket.Conn, ctx *alps.Context) error {
	c, err := ctx.Session.NewIMAP()
	defer c.Close()
	if err != nil {
		return err
	}

	if _, err := c.Select("INBOX", false); err != nil {
		return err
	}

	// create a channel to receive mailbox updates
	updates := make(chan client.Update)
	c.Updates = updates

	// start idling
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- c.Idle(stop, nil)
	}()

	for {
		select {
		case <-done:
			return nil
		case update := <-updates:
			formatted, err := formatUpdate(update)
			if err == nil {
				conn.WriteMessage(websocket.TextMessage, []byte(formatted))
			}
		}
	}
}

func init() {
	p := alps.GoPlugin{Name: "events"}
	alps.RegisterPluginLoader(p.Loader())

	p.GET("/events", func(ctx *alps.Context) error {
		w := ctx.Context.Response().Writer
		r := ctx.Request()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return nil
		}

		go handleWebSocket(conn, ctx)

		return nil
	})
}
