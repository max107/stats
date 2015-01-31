package main

import (
	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/now"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var session *mgo.Session

const SERVER_INFO = "StatsServer"

type BasicServerHeader struct {
	gin.ResponseWriter
	ServerInfo string
}

func (w *BasicServerHeader) WriteHeader(code int) {
	if w.Header().Get("Server") == "" {
		w.Header().Add("Server", w.ServerInfo)
	}

	w.ResponseWriter.WriteHeader(code)
}

func ServerHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &BasicServerHeader{c.Writer, SERVER_INFO}
		c.Writer = writer
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	}
}

func InsertRecord(rec Stats) {
	c := session.DB("stats").C("stats")
	rec.Timestamp = new(Timestamp)
	err := c.Insert(&rec)
	if err != nil {
		log.Fatal(err)
	}
}

func Render(template string, context pongo2.Context) (string, error) {
	tpl, err := pongo2.FromFile(template)
	if err != nil {
		return "", err
	}
	out, err := tpl.Execute(context)
	if err != nil {
		return "", err
	}
	return out, nil
}

func main() {
	log.Printf("%v", time.Now().Local())
	sites := make(map[string]bool)
	sites["localhost:8000"] = true

	var err error
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Timeout:  10 * time.Second,
		Database: "stats",
	}
	session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	r := gin.Default()

	// Global middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.Use(ServerHeader())

	r.OPTIONS("/", func(c *gin.Context) {
		c.String(200, "")
	})

	r.GET("/", func(c *gin.Context) {
		jsDep, _ := ioutil.ReadFile("./static/js/deps.js")
		jsJson, _ := ioutil.ReadFile("./static/js/json.js")
		js, _ := ioutil.ReadFile("./static/js/stats.js")
		c.String(200, string(jsDep)+string(jsJson)+string(js))
	})

	r.POST("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")

		var s Stats
		c.Bind(&s)

		// log.Printf("%s", s.Location.Host)
		// if _, ok := sites[s.Location.Host]; ok {
		InsertRecord(s)
		c.JSON(200, gin.H{"status": true})
		// } else {
		// c.JSON(200, gin.H{"error": "unknown site"})
		// }
	})

	r.GET("/report", func(c *gin.Context) {
		var stats []Stats
		// err = session.DB("stats").C("stats").Find(nil).All(&stats)
		t, _ := now.Parse("2014-01-01")
		log.Printf("%s %s", t, now.EndOfMonth())
		now.FirstDayMonday = true
		match := bson.M{
			"$match": bson.M{
				"timestamp": bson.M{
					"$gte": t, // now.BeginningOfMonth()
					"$lte": now.EndOfMonth(),
				},
			},
		}
		project := bson.M{
			"$project": bson.M{
				"_id": 0,
				"day_of_month": bson.M{
					"$dayOfMonth": "$timestamp",
				},
			},
		}
		group := bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"day_of_month": "$day_of_month",
				},
				"hits": bson.M{"$sum": 1},
			},
		}
		operations := []bson.M{match, project, group}
		pipe := session.DB("stats").C("stats").Pipe(operations)
		err := pipe.All(&stats)
		// err = session.DB("stats").C("stats").Find().Sort("timestamp").All(&stats)
		if err != nil {
			panic(err)
		}

		log.Printf("%v", len(stats))

		out, err := Render("templates/report.html", pongo2.Context{
			"sites": sites,
			"stats": stats,
		})
		if err != nil {
			log.Printf("%v", err)
		}

		c.Data(200, "text/html", []byte(out))
	})

	r.GET("/charts", func(c *gin.Context) {
		out, err := Render("templates/charts.html", pongo2.Context{})
		if err != nil {
			log.Printf("%v", err)
		}

		c.Data(200, "text/html", []byte(out))
	})

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
