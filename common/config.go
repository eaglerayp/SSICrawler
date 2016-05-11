package common

import (
	"Crawler/dao"
	"fmt"
	config "github.com/zpatrick/go-config"
)

const (
	MGO_HOST     = "localhost"
	MGO_PORT     = "27017"
	MGO_MAX_CONN = 250
	MGO_MAX_IDLE = 240
	MGO_USER     = "root"
	MGO_DATABASE = "admin"
	MGO_TIMEOUT  = 300
)

var cfgs *config.Config

func LoadConfig(configFile string) {
	var err error
	tomlProvider := config.NewTOMLFile(configFile)
	mappings := map[string]string{
		"DB_MGO_MAX_CONN": "database.mgo.default.max_conn",
	}
	env := config.NewEnvironment(mappings)
	cfgs = config.NewConfig([]config.Provider{tomlProvider, env})
	err = cfgs.Load()
	if err != nil {
		fmt.Println(err)
	}
}

func GetMgoDBInfo() (ret *dao.DBInfo) {
	var mgoDefaultHost, mgoDefaultPort, mgoDefaultUser, mgoDefaultPassword, mgoDefaultDatabase string
	mgoDefaultHost, _ = cfgs.StringOr("database.mgo.default.host", MGO_HOST)
	mgoDefaultPort, _ = cfgs.StringOr("database.mgo.default.port", MGO_PORT)
	mgoDefaultUser, _ = cfgs.StringOr("database.mgo.default.user", MGO_USER)
	mgoDefaultPassword, _ = cfgs.StringOr("database.mgo.default.password", "")
	mgoDefaultDatabase, _ = cfgs.StringOr("database.mgo.default.database", MGO_DATABASE)
	addr := mgoDefaultHost + ":" + mgoDefaultPort
	ret = dao.NewDBInfo([]string{addr}, mgoDefaultUser, mgoDefaultPassword, mgoDefaultDatabase)
	return
}

// (addrs []string, user, password, dbName string)
