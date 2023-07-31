package irc

import (
	"reflect"
	"testing"
)

func Test_message_GetType(t *testing.T) {
	type fields struct {
		raw         string
		messageType MessageType
	}
	type testStruct struct {
		name   string
		fields fields
		want   MessageType
	}

	// test cases
	tests := []testStruct{
		{
			"Unknown",
			fields{
				raw: ":tmi.twitch.tv 372 justinfan4321 :You are in a maze of twisty passages, all alike.",
			},
			Unknown,
		},
		{
			"Ping",
			fields{
				raw: "PING :tmi.twitch.tv",
			},
			Ping,
		},
		{
			"Reconnect",
			fields{
				raw: ":tmi.twitch.tv RECONNECT",
			},
			Reconnect,
		},
		{
			"Join",
			fields{
				raw: ":justinfan4321!justinfan4321@justinfan4321.tmi.twitch.tv JOIN #sodapoppin",
			},
			Join,
		},
		{
			"Part",
			fields{
				raw: ":justinfan4321!justinfan4321@justinfan4321.tmi.twitch.tv PART #sodapoppin",
			},
			Part,
		},
		{
			"PrivMessage",
			fields{
				raw: ":fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin :sodaHmm Did you know you get a free subscription when you link your Amazon Prime account with Twitch? You can sub to Chance for free and spam emotes all day! sodaH sodaL - https://gaming.amazon.com\n",
			},
			PrivMessage,
		},
		{
			"PrivMessage_Tags",
			fields{
				raw: "@badge-info=subscriber/57;badges=moderator/1,subscriber/3054,partner/1;color=#1976D2;display-name=Fossabot;emotes=;first-msg=0;flags=;id=23ebb86b-f9fa-47b8-893c-708587661afc;mod=1;returning-chatter=0;room-id=26301881;subscriber=1;tmi-sent-ts=1690815698066;turbo=0;user-id=237719657;user-type=mod :fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin : ( ° ͜ʖ͡°)╭∩╮",
			},
			PrivMessage,
		},
		{
			"Cap",
			fields{
				raw: ":tmi.twitch.tv CAP * ACK :twitch.tv/tags",
			},
			Cap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Message{
				raw:         tt.fields.raw,
				messageType: tt.fields.messageType,
			}
			if got := m.GetType(); got != tt.want {
				t.Errorf("GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMessage(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    *Message
		wantErr bool
	}{
		{
			name: "PrivMessage",
			args: args{data: "@badge-info=subscriber/57;badges=moderator/1,subscriber/3054,partner/1;color=#1976D2;display-name=Fossabot;emotes=;first-msg=0;flags=;id=23ebb86b-f9fa-47b8-893c-708587661afc;mod=1;returning-chatter=0;room-id=26301881;subscriber=1;tmi-sent-ts=1690815698066;turbo=0;user-id=237719657;user-type=mod :fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin : ( ° ͜ʖ͡°)╭∩╮"},
			want: &Message{
				raw:         "@badge-info=subscriber/57;badges=moderator/1,subscriber/3054,partner/1;color=#1976D2;display-name=Fossabot;emotes=;first-msg=0;flags=;id=23ebb86b-f9fa-47b8-893c-708587661afc;mod=1;returning-chatter=0;room-id=26301881;subscriber=1;tmi-sent-ts=1690815698066;turbo=0;user-id=237719657;user-type=mod :fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin : ( ° ͜ʖ͡°)╭∩╮",
				messageType: PrivMessage,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessage(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
