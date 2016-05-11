package dao

import (
	// "encoding/json"
	"gopkg.in/mgo.v2"
	"log"
)

const Database = "SSI"

type DBInfo struct {
	Addrs    []string
	User     string
	Password string
	Database string
}

type Resource struct {
	session *mgo.Session
}

func NewDBInfo(addrs []string, user, password, dbName string) *DBInfo {
	return &DBInfo{
		Addrs:    addrs,
		User:     user,
		Password: password,
		Database: dbName,
	}
}

func (d *Resource) GetSession() *mgo.Session {
	return d.session.Copy()
}

func ConnectMongo(dbi *DBInfo) (*Resource, error) {
	mgoDialInfo := &mgo.DialInfo{
		Addrs:    dbi.Addrs,
		Username: dbi.User,
		Password: dbi.Password,
		Database: dbi.Database,
	}
	mgoSession, err := mgo.DialWithInfo(mgoDialInfo)
	if err != nil {
		log.Println(err)
		return &Resource{}, err
	}
	return &Resource{session: mgoSession}, nil
}
