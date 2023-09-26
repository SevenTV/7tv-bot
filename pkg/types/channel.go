package types

import "time"

// Channel is the struct we use to decode data from mongo
// TODO: keep track of username changes/history?
type Channel struct {
	ID        int64     `bson:"user_id" json:"user_id"`
	Flags     uint32    `bson:"flags,omitempty" json:"flags"`
	Username  string    `bson:"username" json:"username"`
	Platform  string    `bson:"platform" json:"platform"`
	Weight    int       `bson:"weight" json:"weight"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
