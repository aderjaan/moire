package db

import (
	"gopkg.in/mgo.v2/bson"
)

type BaseInterface interface {
	PrepareSave()
}

type BaseModel struct {
	Id         bson.ObjectId `json:"_id,omitempty" bson:"_id" required:"true"`
	CreatedOn  int64         `json:"created_on" bson:"created_on" required:"true"`
	ModifiedOn int64         `json:"updated_on" bson:"updated_on" required:"true"`
}
