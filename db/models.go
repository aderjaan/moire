package db

import "gopkg.in/mgo.v2/bson"

const (
	ASSET   = "asset"
	PENDING = "pending"
	READY   = "ready"
	LOST    = "lost"
)

type Asset struct {
	Id            bson.ObjectId `json:"_id,omitempty" bson:"_id",omitempty`
	CreatedOn     int64         `json:"created_on" bson:"created_on"`
	Name          string        `json:"name" bson:"name"`
	FileType      string        `json:"file_type" bson:"file_type"`
	MimeType      string        `json:"mime_type" bson:"mime_type"`
	Bucket        string        `json:"bucket" bson:"bucket"`
	Path          string        `json:"path" bson:"path"`
	ThumbnailPath string        `json:"thumbnail_path" bson:"thumbnail_path"`
	Status        string        `json:"status" bson:"status"`
	History       []string      `json:"-" bson:"history"`
}
