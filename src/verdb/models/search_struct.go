package models

import "gopkg.in/mgo.v2/bson"

type SearchStruct struct {
	Query     bson.M   `json:"query" binding:"required"`
	Selection bson.M   `json:"selection"`
	Sort      []string `json:"sort"`
	Limit     int      `json:"limit"`
}
