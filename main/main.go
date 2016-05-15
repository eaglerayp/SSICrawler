// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"Crawler/common"
	"Crawler/dao"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"time"
	// "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const fxcmUrl = "wss://streamssi.fxcorporate.com/streamssi/endpoint/EURUSD,GBPUSD,AUDUSD,USDCAD,NZDUSD,XAUUSD"

func parseSSIData(c *websocket.Conn) ([]map[string]interface{}, error) {
	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read error:", err)
		return nil, err
	}
	str := string(message[:])
	// log.Printf("recv: %s", str)
	str = str[4 : len(str)-1]

	var ssiData map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(str), &ssiData); err != nil {
		panic(err)
	}
	return ssiData["SSI"], nil
}

func saveDataToMongo(d *dao.Resource, ssiArray []map[string]interface{}) {
	timeString := time.Now().Format(time.RFC3339)
	session := d.GetSession()

	for _, ssi := range ssiArray {
		// use time.Now replace the SSI given time. lazy to parse the format
		// insert data to mongodb
		collection := session.DB(dao.Database).C(ssi["Symbol"].(string))
		data := bson.M{"SSIHistOrders": ssi["SSIHistOrders"], "Time": timeString}
		err := collection.Insert(data)
		if err != nil {
			log.Println("SSI ", ssi["Symbol"], ":", ssi["SSIHistOrders"], "; time:", timeString)
			log.Println("insert error:", err)
		}

	}
	session.Close()
}

func main() {
	log.Println("Start: FXCM-SSI-WScrawler v1.0")
	// config for mongodb
	var cfgFile string
	flag.StringVar(&cfgFile, "config", "./config.toml", "Path to Config File")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <command> \n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	common.LoadConfig(cfgFile)
	// setup connection to mongodb
	d, err := dao.ConnectMongo(common.GetMgoDBInfo())
	if err != nil {
		panic(err)
	}
	log.Println("Connecting to mongo successfully, start crawling data!")
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})
	// main save data rule
	go func() {
		for {
			defer close(done)
			c, _, err := websocket.DefaultDialer.Dial(fxcmUrl, nil)
			if err != nil {
				log.Fatal("dial:", err)
			}
			ssiArray, err := parseSSIData(c)
			if err != nil {
				log.Fatal("Parse SSIData Error:", err)
			}

			// check data changing then save data to prevent weekend useless data
			isSame := true
			oldData := ssiArray[0]["SSIHistOrders"].(string)
			for isSame == true {
				time.Sleep(1 * time.Second)
				ssiArray, _ := parseSSIData(c)
				isSame = oldData == ssiArray[0]["SSIHistOrders"].(string)
			}

			saveDataToMongo(d, ssiArray)
			c.Close()
			log.Println("Save SSI Data success, sleep 1 hour!")
			time.Sleep(1 * time.Hour)
		}
	}()

	// main controller
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
