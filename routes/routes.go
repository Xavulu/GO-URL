package routes

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"url-shortener/db"
	"url-shortener/shortener"
	"url-shortener/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func submitURL(c *gin.Context) {
	var input utils.InputURL
	var out utils.PostResponse
	fmt.Println(out)
	if err := c.ShouldBindBodyWith(&input, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
			"message": "you did not enter a url",
			"error":   "your input should be a url"})
		return
	}
	if utils.IsURL(input.URL) == false {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
			"message": "invalid url format, should be : https://www.example.com",
			"error":   input.URL})
		return
	}
	exists, fail := db.URLExists(input.URL)
	if fail != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
			"message": "Error checking for url :(",
			"error":   fail})
		return
	}
	if exists == true {
		short, notFound := db.FetchShort(input.URL)
		if notFound != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
				"message": fmt.Sprintf("Error fetching short url for %s :(", input.URL),
				"error":   notFound})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
			"message": fmt.Sprintf("%s already exists at http://localhost:8080/o/%s", input.URL, short),
			"error":   "attempted duplicate url entry"})
		return
	}
	out.Origin = input.URL
	out.Entered = time.Now()
	conn := db.InitDb()
	err := conn.QueryRow(context.Background(), db.InsertURL, out.Origin, out.Entered).Scan(&out.ID)
	if err != nil {
		defer conn.Close()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Error inserting url :(",
			"error":   err})
		return
	}
	tempshort := shortener.Encoder(int64(out.ID))
	i, err := conn.Exec(context.Background(), db.InsertShort, tempshort, out.ID)
	if err != nil {
		defer conn.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
			"message": "Could not create shortened url :(",
			"error":   err})
		return
	}
	fmt.Println(i)
	out.Shortened = fmt.Sprintf("http://localhost:8080/o/%s", tempshort)
	defer conn.Close()
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK,
		"message": out})
	return
}

func sendTo(c *gin.Context) {
	short := c.Param("id")
	if utils.IsValidShort(short) == true {
		urlID := shortener.Decoder(short)
		var url string
		conn := db.InitDb()
		err := conn.QueryRow(context.Background(), db.GetURL, urlID).Scan(&url)
		if err != nil {
			defer conn.Close()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "this url does not exist in the database",
				"error":   "bad request"})
			return
		}
		visited, err := conn.Exec(context.Background(), db.UpdateVisits, urlID)
		if err != nil {
			defer conn.Close()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    err,
				"message": "could not update url information",
				"error":   "internal server error"})
			return
		}
		stats, err := conn.Exec(context.Background(), db.TrackStats, short, time.Now())
		if err != nil {
			defer conn.Close()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "could not update redirection statistics",
				"error":   "internal server error"})
			return
		}
		fmt.Println(visited, stats) //so go wont scream at me, same reason why these had names instead of just _, err
		defer conn.Close()
		c.Redirect(http.StatusMovedPermanently, url)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
		"message": "you entered an invalid shorturl",
		"error":   "invalid"})
	return
}

func viewURL(c *gin.Context) {
	short := c.Param("id")
	if utils.IsValidShort(short) == true {
		urlID := shortener.Decoder(short)
		var url string
		conn := db.InitDb()
		err := conn.QueryRow(context.Background(), db.GetURL, urlID).Scan(&url)
		defer conn.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "this url does not exist in the database",
				"error":   err})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": fmt.Sprintf("http://localhost:8080/o/%s links to %s", short, url)})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
		"message": "you entered an invalid shorturl",
		"error":   "invalid"})
	return
}

func dataDashBoard(c *gin.Context) {
	short := c.Param("id")
	if utils.IsValidShort(short) == true {
		search, err := db.ShortExists(short)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
				"message": "could not verify existence of shorturl",
				"error":   "internal server error"})
			return
		}
		if search == true {
			times, err := db.URLVisits(short)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
					"message": "could not fetch url visit statistics",
					"error":   "internal server error"})
				return
			}
			conn := db.InitDb()
			var stats utils.Stats
			fail := conn.QueryRow(context.Background(), db.GetURL, shortener.Decoder(short)).Scan(&stats.Baseurl)
			if fail != nil {
				defer conn.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
					"message": "could not fetch url visit statistics",
					"error":   "internal server error"})
				return
			}
			err2 := conn.QueryRow(context.Background(), db.GetStatus, shortener.Decoder(short)).Scan(&stats.Visited, &stats.Visitcount)
			if err2 != nil {
				defer conn.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError,
					"message": "could not fetch url visit statistics",
					"error":   "internal server error"})
				return
			}
			stats.Shorturl = fmt.Sprintf("http://localhost:8080/o/%s", short)
			stats.Uniquevisits = times
			defer conn.Close()
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK,
				"message": stats})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
			"message": fmt.Sprintf("http://localhost:8080/0/%s does not exist in the database", short),
			"error":   "invalid"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest,
		"message": "you entered an invalid shorturl",
		"error":   "invalid"})
	return
}

//StartRouter initializes gin and its routing, starts server
func StartRouter() {
	myfile, err := os.Create("short_url_server.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR]: %v\n", err)
		os.Exit(1)
	}
	gin.DefaultWriter = io.MultiWriter(myfile, os.Stdout)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(gin.Logger())
	router.POST("/entry", submitURL)
	router.GET("/o/:id", sendTo) //the o has to be there, comprimising url shortness, because of gin's bizzare routing method errors
	router.GET("/view/:id", viewURL)
	router.GET("/stats/:id", dataDashBoard)
	router.Run()
}
