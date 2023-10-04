package emotes

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Collections struct {
	GlobalStats *mongo.Collection
}

var collections Collections

func SetCollections(globalStats *mongo.Collection) {
	collections.GlobalStats = globalStats
}
