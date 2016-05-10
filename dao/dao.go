package dao

import (
	// "encoding/json"
	"gopkg.in/mgo.v2"
	"log"
)

type Resource struct {
	session *mgo.Session
}

func (d *Resource) GetSession() *mgo.Session {
	return d.session.Copy()
}

func ConnectMongo() (*Resource, error) {
	mgoDialInfo := &mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Username: "test",
		Password: "test",
		Database: "test",
	}
	mgoSession, err := mgo.DialWithInfo(mgoDialInfo)
	if err != nil {
		log.Println(err)
		return &Resource{}, err
	}
	return &Resource{session: mgoSession}, nil
}
