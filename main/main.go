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
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	// // log.Printf("connecting to %s", u.String())
	fxcmUrl := "wss://streamssi.fxcorporate.com/streamssi/endpoint/EURUSD,GBPUSD,AUDUSD,USDCAD,NZDUSD,XAUUSD"
	c, _, err := websocket.DefaultDialer.Dial(fxcmUrl, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			str := string(message[:])
			// log.Printf("recv: %s", str)
			str = str[4 : len(str)-1]
			// log.Printf("str after: %s", str)
			// ssiData format = {SSI: [{"Symbol":?,"Time":?,},{},{}]
			// }
			var ssiData map[string][]map[string]interface{}
			if err := json.Unmarshal([]byte(str), &ssiData); err != nil {
				panic(err)
			}
			ssiArray := ssiData["SSI"]
			timeString := time.Now().Format(time.RFC3339)
			session := d.GetSession()

			for _, ssi := range ssiArray {
				// use time.Now replace the SSI given time. lazy to parse the format
				// insert data to mongodb
				collection := session.DB(dao.Database).C(ssi["Symbol"].(string))
				data := bson.M{"SSIHistOrders": ssi["SSIHistOrders"], "Time": timeString}
				err = collection.Insert(data)
				if err != nil {
					log.Println("SSI ", ssi["Symbol"], ":", ssi["SSIHistOrders"], "; time:", timeString)
					log.Println("insert error:", err)
				}

			}
			session.Close()

			// sleep
			// time.Sleep(10 * time.Second)
			// log.Println("Sleep 10 Second")
			log.Println("Sleep 1 hour")
			time.Sleep(1 * time.Hour)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}
