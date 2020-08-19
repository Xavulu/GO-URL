package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	//sorry if this is ugly its just more readable to me :(
	dropShort      string = "DROP TABLE IF EXISTS shortener;"
	dropVisit      string = "DROP TABLE IF EXISTS visits;"
	shortenerTable string = `CREATE TABLE IF NOT EXISTS 
	shortener (id SERIAL PRIMARY KEY NOT NULL, 
		url VARCHAR(255) CONSTRAINT unique_url UNIQUE NOT NULL, 
		encoded VARCHAR, 
		visited BOOLEAN DEFAULT FALSE, 
		count INTEGER DEFAULT 0, 
		created_at TIMESTAMP NOT NULL);`
	visitTable string = `CREATE TABLE IF NOT EXISTS 
	visits(id SERIAL PRIMARY KEY NOT NULL, 
		ref VARCHAR NOT NULL, 
		visited TIMESTAMP NOT NULL);`
)

var (
	//InsertURL inserts a url
	InsertURL string = `INSERT INTO shortener(url, created_at) 
	VALUES ($1, $2) RETURNING id;`
	//CheckExists checks if a url exists in the database already
	CheckExists string = `SELECT EXISTS(SELECT 1 FROM shortener WHERE url = $1);`
	//IfExists returns shorturl for url that already exists
	IfExists string = `SELECT encoded FROM shortener WHERE url = $1;` //RETURNING encoded;`
	//InsertShort inserts the base62 encoded string
	InsertShort string = `UPDATE shortener SET encoded = $1 WHERE id = $2;`
	//GetURL fetches the url based off the id
	GetURL string = `SELECT url FROM shortener where id = $1;`
	//UpdateVisits updates the visited section to true and increments the visit count
	UpdateVisits string = `UPDATE shortener SET visited = true, count = count + 1 WHERE id = $1;`
	//TrackStats adds visitation statistics to the visits tables
	TrackStats string = `INSERT INTO visits(ref, visited) VALUES ($1, $2);`
	//CheckShort is used to check if a shorturl exists before trying to get the visit statistics
	CheckShort string = `SELECT EXISTS(SELECT 1 FROM shortener WHERE encoded = $1);`
	//GetTimes fetches the visit statistics for a chosen url
	GetTimes string = `SELECT visited FROM visits WHERE ref = $1;`
	//GetStatus fetches visited (boolean) and count(int) from the database
	GetStatus string = `SELECT visited, count FROM shortener WHERE id = $1;`
)

//InitDb creates a connection to our postgres db and returns the connection , has to be closed after
func InitDb() *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	connPool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return connPool
}

//populateDb populates our databse with the appropiate tables i.e id, url, base64 encoding, visited, visit count
func populateDb(db *pgxpool.Pool) {
	check1, err := db.Exec(context.Background(), dropShort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to drop existing shortener table: %v\n", err)
		os.Exit(1)
	}
	check2, err := db.Exec(context.Background(), dropVisit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to drop existing visits table: %v\n", err)
		os.Exit(1)
	}
	insert1, err := db.Exec(context.Background(), shortenerTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create shortener table: %v\n", err)
		os.Exit(1)
	}
	insert2, err := db.Exec(context.Background(), visitTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create visits table: %v\n", err)
	}
	fmt.Println(check1, check2, insert1, insert2) // just here so go doesnt scream at me for unused variables
}

//StartDb makes the initial db connection after the postgres database is up and running by calling Initdb and populatedb
func StartDb() {
	time.Sleep(10000 * time.Millisecond)
	/*since docker-compose is being used it was necessary to make the initial connection take longer as the go app
	can run before postgres is ready to accept connections*/
	conn := InitDb()
	populateDb(conn)
	defer conn.Close()
}

//URLExists checks if a url already exists in the database to make sure no duplicates are entered
func URLExists(url string) (bool, error) {
	var exists bool
	conn := InitDb()
	err := conn.QueryRow(context.Background(), CheckExists, url).Scan(&exists)
	defer conn.Close()
	return exists, err
}

//FetchShort is used together with URLExists to return the encoded string for a url that already exists in the database
func FetchShort(url string) (string, error) {
	var short string
	conn := InitDb()
	err := conn.QueryRow(context.Background(), IfExists, url).Scan(&short)
	defer conn.Close()
	return short, err
}

//ShortExists checks to see if a short url actually exists in the database
func ShortExists(short string) (bool, error) {
	var exists bool
	conn := InitDb()
	err := conn.QueryRow(context.Background(), CheckShort, short).Scan(&exists)
	defer conn.Close()
	return exists, err
}

//URLVisits collects all the visit times from the database and retursn them in an array alongside an error
func URLVisits(shorturl string) ([]time.Time, error) {
	var visits []time.Time
	//var visit time.Time
	conn := InitDb()
	times, err := conn.Query(context.Background(), GetTimes, shorturl)
	defer conn.Close()
	for times.Next() {
		var visit time.Time
		err := times.Scan(&visit)
		if err != nil {
			return visits, err
		}
		visits = append(visits, visit)
	}
	return visits, err
}
