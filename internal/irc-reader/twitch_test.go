package irc_reader

import "testing"

func Test_parseChannel(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "privmsg+tags",
			args: args{
				in: "@badge-info=subscriber/57;badges=moderator/1,subscriber/3054,partner/1;color=#1976D2;display-name=Fossabot;emotes=;first-msg=0;flags=;id=23ebb86b-f9fa-47b8-893c-708587661afc;mod=1;returning-chatter=0;room-id=26301881;subscriber=1;tmi-sent-ts=1690815698066;turbo=0;user-id=237719657;user-type=mod :fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin : ( ° ͜ʖ͡°)╭∩╮",
			},
			want: "sodapoppin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseChannel(tt.args.in); got != tt.want {
				t.Errorf("parseChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMessageId(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "privmsg+tags",
			args: args{
				in: "@badge-info=subscriber/57;badges=moderator/1,subscriber/3054,partner/1;color=#1976D2;display-name=Fossabot;emotes=;first-msg=0;flags=;id=23ebb86b-f9fa-47b8-893c-708587661afc;mod=1;returning-chatter=0;room-id=26301881;subscriber=1;tmi-sent-ts=1690815698066;turbo=0;user-id=237719657;user-type=mod :fossabot!fossabot@fossabot.tmi.twitch.tv PRIVMSG #sodapoppin : ( ° ͜ʖ͡°)╭∩╮",
			},
			want: "23ebb86b-f9fa-47b8-893c-708587661afc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMessageId(tt.args.in); got != tt.want {
				t.Errorf("parseMessageId() = %v, want %v", got, tt.want)
			}
		})
	}
}
