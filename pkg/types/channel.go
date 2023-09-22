package types

import "time"

// Channel is the struct we use to decode data from mongo
// TODO: keep track of username changes/history?
type Channel struct {
	ID        int64     `bson:"user_id"`
	Flags     uint32    `bson:"flags,omitempty"`
	Username  string    `bson:"username"`
	Platform  string    `bson:"platform"`
	Weight    int       `bson:"weight"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}
