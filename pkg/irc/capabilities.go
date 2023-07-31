package irc

const (
	// CapTags adds additional metadata to the command and membership messages. Such as user ID, display name, badges etc.
	CapTags = "twitch.tv/tags"
	// CapCommands lets your bot send PRIVMSG messages that include Twitch chat commands and receive Twitch-specific IRC messages.
	CapCommands = "twitch.tv/commands"
	// CapMembership lets your bot receive JOIN and PART messages when users join and leave the chat room.
	CapMembership = "twitch.tv/membership"
)
