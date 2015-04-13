package db

import "gopkg.in/mgo.v2/bson"

const (
	ASSET   = "asset"
	PENDING = "pending"
	READY   = "ready"
	LOST    = "lost"
)

type Asset struct {
	Id           bson.ObjectId `json:"_id,omitempty" bson:"_id",omitempty`
	CreatedOn    int64         `json:"created_on" bson:"created_on"`
	Name         string        `json:"name" bson:"name"`
	MimeType     string        `json:"mime_type" bson:"mime_type"`
	Url          string        `json:"url" bson:"url"`
	ThumbnailUrl string        `json:"thumbnail_url" bson:"thumbnail_url"`
	Status       string        `json:"status" bson:"status"`
	History      []string      `json:"-" bson:"history"`
}
