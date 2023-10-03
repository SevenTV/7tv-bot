package emotes

type EmoteCount struct {
	Name    string `bson:"name"`
	EmoteID string `bson:"emoteID"`
	Flags   int32  `bson:"flags"`
	URL     string `bson:"url"`
	Count   int    `bson:"count"`
}
