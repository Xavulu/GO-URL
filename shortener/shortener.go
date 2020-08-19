package shortener

const (
	base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	len62  = int64(len(base62))
)

//Encoder receives the uid from the database entry and base62 encodes it
func Encoder(idnum int64) string {
	if idnum == 0 {
		return string(base62[0])
	}
	encoded := ""
	for ; idnum > 0; idnum = idnum / len62 {
		encoded = string(base62[idnum%len62]) + encoded
	}
	return encoded
} // encoding method is taken from https://github.com/douglasmakey/ursho/blob/master/base62/base62.go

/*Decoder well, decodes the base642 string from the end of the
get request and returns the matching uid to check against the database */
func Decoder(idstr string) int64 {
	var uid int64
	shortURL := []byte(idstr)
	for i := 0; i < len([]byte(idstr)); i++ {
		if 'a' <= shortURL[i] && shortURL[i] <= 'z' {
			uid = uid*62 + int64(shortURL[i]) - 'a'
		}
		if 'A' <= shortURL[i] && shortURL[i] <= 'Z' {
			uid = uid*62 + int64(shortURL[i]) - 'A' + 26
		}
		if '0' <= shortURL[i] && shortURL[i] <= '9' {
			uid = uid*62 + int64(shortURL[i]) - '0' + 52
		}
	}
	return uid
} //based off the c++ method on geeks for geeks
