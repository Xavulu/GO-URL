package utils

import (
	"net/url"
	"regexp"
	"time"
)

///////////////////////

//IsValidShort uses a regular expression to check if the shorturl contains valid characters (base62 only)
var (
	IsValidShort = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString
)

//InputURL for entering url
type InputURL struct {
	URL string `json:"url" binding:"required"`
}

//Encoded for entering the base62 encoded id
type Encoded struct {
	Short string `json:"short_url" binding:"required"`
}

//PostResponse is a struct for creating response to POST
type PostResponse struct {
	ID        int64     `json:"id"`
	Origin    string    `json:"original_url"`
	Shortened string    `json:"short_url"`
	Entered   time.Time `json:"entry_date"`
}

//Stats is for creating a json response for the GET /view/:id route
type Stats struct {
	Baseurl      string      `json:"original_url"`
	Shorturl     string      `json:"short_url"`
	Visited      bool        `json:"visited"`
	Visitcount   int         `json:"visit_count"`
	Uniquevisits []time.Time `json:"visit_times"`
}

//////////////////////

/*IsURL checks if a valid url (ex: http://example.com) is being passed
and returns true or false, this is necessary to make sure invalid urls are
not passed into the database. Method taken from:
https://stackoverflow.com/questions/31480710/validate-url-with-standard-package-in-go
*/
func IsURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && u.Scheme != "" && u.Host != ""
}
