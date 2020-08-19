package main

import (
	"url-shortener/db"
	"url-shortener/routes"

	_ "github.com/gin-gonic/gin"
)

func main() {
	/*
		set databse url in this format in root directories dockerfile
		DATABASE_URL=postgres://user:password@host:port/database
	*/
	db.StartDb()
	routes.StartRouter()
}
