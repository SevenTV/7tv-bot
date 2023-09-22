package manager

import (
	"reflect"
	"testing"

	"github.com/seventv/7tv-bot/pkg/irc"
)

func Test_connection_handleMessages(t *testing.T) {
	joinMsg, _ := irc.ParseMessage(":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv JOIN #forsen")
	partMsg, _ := irc.ParseMessage(":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv PART #forsen")
	type fields struct {
		client    *irc.Client
		channels  []*IRCChannel
		capacity  int
		onMessage func(msg *irc.Message, err error)
	}
	type args struct {
		msg *irc.Message
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*IRCChannel
	}{
		{
			name: "JOIN",
			fields: fields{
				channels: []*IRCChannel{
					{Name: "forsen"},
					{Name: "sodapoppin"},
				},
				onMessage: func(msg *irc.Message, err error) {},
			},
			args: args{
				msg: joinMsg,
				err: nil,
			},
			want: []*IRCChannel{
				{
					Name:     "forsen",
					isJoined: true,
				},
				{
					Name:     "sodapoppin",
					isJoined: false,
				},
			},
		},
		{
			name: "PART",
			fields: fields{
				channels: []*IRCChannel{
					{
						Name:     "forsen",
						isJoined: true,
					},
					{
						Name:     "sodapoppin",
						isJoined: true,
					},
				},
				onMessage: func(msg *irc.Message, err error) {},
			},
			args: args{
				msg: partMsg,
				err: nil,
			},
			want: []*IRCChannel{
				{
					Name:     "forsen",
					isJoined: false,
				},
				{
					Name:     "sodapoppin",
					isJoined: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connection{
				client:    tt.fields.client,
				channels:  tt.fields.channels,
				capacity:  tt.fields.capacity,
				onMessage: tt.fields.onMessage,
			}
			c.handleMessages(tt.args.msg, tt.args.err)
			if !reflect.DeepEqual(c.channels, tt.want) {
				t.Errorf("handleMessages() %v, result didn't equal what we expected!", tt.name)
			}
		})
	}
}

func Test_parseChannels(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "JOIN",
			args: args{
				data: ":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv JOIN #sodapoppin",
			},
			want: []string{"sodapoppin"},
		},
		{
			name: "PART",
			args: args{
				data: ":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv PART #sodapoppin",
			},
			want: []string{"sodapoppin"},
		},
		{
			name: "PARTMany",
			args: args{
				data: ":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv PART #sodapoppin,#forsen",
			},
			want: []string{"sodapoppin", "forsen"},
		},
		{
			name: "PartialMessage",
			args: args{
				data: ":justinfan77777!justinfan77777@justinfan77777.tmi.twitch.tv PART",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseChannels(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePingPayload(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "PING",
			args: args{data: "PING :tmi.twitch.tv"},
			want: " :tmi.twitch.tv",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePingPayload(tt.args.data); got != tt.want {
				t.Errorf("parsePingPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
