package aggregator

import (
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/seventv/7tv-bot/pkg/irc"
)

func (s *Service) handleMessage(natsMsg *nats.Msg) error {
	msg, err := parseMessage(natsMsg.Data)
	if err != nil {
		return err
	}

	return nil
}

type Message struct {
	Sender struct {
		ID       string
		Username string
	}
	Room struct {
		Username string
		ID       string
	}
	Type         string
	MessageWords []string
	Raw          string
}

func parseMessage(data []byte) (*Message, error) {
	rawMessage := string(data)
	msg := &Message{Raw: rawMessage}

	var i int
	split := strings.Split(rawMessage, " ")

	if strings.HasPrefix(rawMessage, "@") {
		err := parseTags(msg, split[i])
		if err != nil {
			return nil, err
		}
		i++
	}

	if i >= len(split) {
		return nil, irc.ErrPartialMessage
	}

	if strings.HasPrefix(split[i], ":") {
		err := parseSender(msg, split[i])
		if err != nil {
			return nil, err
		}
		i++
	}

	if i >= len(split) {
		return nil, irc.ErrPartialMessage
	}

	msg.Type = split[i]
	i++

	if i >= len(split) {
		return nil, irc.ErrPartialMessage
	}

	msg.Room.Username = strings.TrimPrefix(split[i], "#")
	i++

	if i >= len(split) {
		return nil, irc.ErrPartialMessage
	}

	msg.MessageWords = split[i:]
	msg.MessageWords[0] = strings.TrimPrefix(msg.MessageWords[0], ":")

	return msg, nil
}

func parseTags(msg *Message, tags string) error {
	split := strings.Split(msg.Raw, ";")
	for _, substr := range split {
		key, value, ok := strings.Cut(substr, "=")
		if !ok {
			continue
		}
		switch key {
		case "user-id":
			msg.Sender.ID = value
		case "room-id":
			msg.Room.ID = value
		}
	}
	return nil
}

func parseSender(msg *Message, sender string) error {
	trim := strings.TrimLeft(sender, ":")
	before, _, ok := strings.Cut(trim, "!")
	if !ok {
		return irc.ErrPartialMessage
	}
	msg.Sender.Username = before
	return nil
}
