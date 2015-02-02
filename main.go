package main

// WARNING! Remember, dont change day_of_month.

import (
	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/now"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

var session *mgo.Session

const MONGO_DB = "stats"
const MONGO_COLLECTION = "stats2"
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

func InsertRecord(rec Stats) error {
	c := session.DB(MONGO_DB).C(MONGO_COLLECTION)
	rec.Timestamp = Timestamp(time.Now().UTC())
	return c.Insert(&rec)
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

func FindByMonth(host string, gte, lte time.Time) ([]StatsMonthResult, error) {
	var results []StatsMonthResult

	now.FirstDayMonday = true
	match := bson.M{
		"$match": bson.M{
			"host": host,
			"timestamp": bson.M{
				"$gte": gte,
				"$lte": lte,
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
	pipe := session.DB(MONGO_DB).C(MONGO_COLLECTION).Pipe(operations)
	err := pipe.All(&results)
	// err = session.DB("stats").C("stats").Find().Sort("timestamp").All(&stats)
	if err != nil {
		return results, err
	}

	return results, nil
}

func main() {
	sites := make(map[string]bool)
	sites["studio107.ru"] = true

	var err error
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Timeout:  10 * time.Second,
		Database: MONGO_DB,
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

		if _, ok := sites[s.Location.Host]; ok {
			err := InsertRecord(s)
			if err != nil {
				panic(err)
			}
			c.JSON(200, gin.H{"status": true})
		} else {
			c.JSON(200, gin.H{"error": "unknown site"})
		}
	})

	r.GET("/report", func(c *gin.Context) {
		out, err := Render("templates/sites.html", pongo2.Context{
			"sites": sites,
		})
		if err != nil {
			panic(err)
		}

		c.Data(200, "text/html", []byte(out))
	})

	r.GET("/report/:host", func(c *gin.Context) {
		host := c.Params.ByName("host")
		if _, ok := sites[host]; !ok {
			c.String(404, "Page not found")
			return
		}
		results, err := FindByMonth(host, now.BeginningOfMonth().UTC(), now.EndOfMonth().UTC())
		if err != nil {
			panic(err)
		}

		// TODO refact
		var chartData = make(map[int]int, now.EndOfMonth().Day())
		for i := 1; i < now.EndOfMonth().Day()+1; i++ {
			var complete = false
			for _, r := range results {
				if i == r.Day() {
					chartData[i] = r.Hits
					complete = true
					break
				}
			}
			if complete == false {
				chartData[i] = 0
			}
		}

		var keys []int
		for k := range chartData {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		// TODO refact

		out, err := Render("templates/report.html", pongo2.Context{
			"chartData": chartData,
			"chartKeys": keys,
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
