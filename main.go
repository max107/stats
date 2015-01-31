package main

import (
	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
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
	sites := make(map[string]bool)
	sites["localhost:8000"] = true

	var err error
	session, err = mgo.Dial("localhost")
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
		jsDep, _ := ioutil.ReadFile("./deps.js")
		jsJson, _ := ioutil.ReadFile("./json.js")
		js, _ := ioutil.ReadFile("./stats.js")
		c.String(200, string(jsDep)+string(jsJson)+string(js))
	})

	r.POST("/", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")

		var s Stats
		c.Bind(&s)

		if _, ok := sites[s.Location.Host]; ok {
			InsertRecord(s)
			c.JSON(200, gin.H{"status": true})
		} else {
			c.JSON(200, gin.H{"error": "unknown site"})
		}
	})

	r.GET("/report", func(c *gin.Context) {
		var stats []Stats
		err = session.DB("stats").C("stats").Find(nil).All(&stats)
		if err != nil {
			panic(err)
		}

		out, err := Render("report.html", pongo2.Context{
			"sites": sites,
			"stats": stats,
		})
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
