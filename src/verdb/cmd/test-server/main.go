package main

import "gopkg.in/mgo.v2"

type Person struct {
	NameTitle string `json:"nameTitle" bson:"nameTitle"`
	Age       int
}

func main() {
	sess, _ := mgo.Dial("localhost")
	defer sess.Close()

	sess.DB("test").C("people").Insert(Person{"Alex", 28})
}
