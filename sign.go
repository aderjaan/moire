package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/url"
	"strings"
	"time"
)

var TOKEN_MAP = map[string]string{}

func getValues(uri *url.URL) url.Values {
	return uri.Query()
}

func getSecretKey(publicKey string) string {
	secret, ok := TOKEN_MAP[publicKey]
	if !ok {
		panic("Invalid publicKey: " + publicKey)
	}

	return secret
}

func getPublicKey(uri *url.URL) string {
	values := getValues(uri)
	publicKey, ok := values["access_token"]
	if !ok || len(publicKey) < 0 {
		panic("Missing publicKey in request")
	}

	return publicKey[0]
}

func isTimestampValid(uri *url.URL) bool {
	values := getValues(uri)
	signed_on, ok := values["timestamp"]
	if !ok || len(signed_on) < 0 {
		panic("Missing timestamp in request")
	}

	timestamp, err := time.Parse(time.RFC3339, signed_on[0])
	if err != nil {
		panic(err)
	}

	current_time := time.Now()

	max_time_skew := current_time.Add(5 * time.Minute)
	max_time_offset := current_time.Add(-5 * time.Minute)

	log.Println("Current:", current_time)
	log.Println("Timestamp:", timestamp)
	log.Println("Skew", max_time_skew)
	log.Println("Offset", max_time_offset)

	if timestamp.Sub(max_time_skew) > 0 {
		panic("Timestamp max skew validation error")
	}

	if timestamp.Sub(max_time_offset) < 0 {
		panic("Timestamp max offset validation error")
	}

	return true
}

func canonicalQuery(uri *url.URL) string {
	values := getValues(uri)
	sorted := values.Encode()
	escaped := strings.Replace(sorted, "+", "%20", -1)
	return escaped
}

func canonicalPath(uri *url.URL) string {
	path := uri.Opaque
	if path != "" {
		path = "/" + strings.Join(strings.Split(path, "/")[3:], "/")
	} else {
		path = uri.Path
	}

	if path == "" {
		path = "/"
	}

	return path
}

func makeHmac512(message, secret string) []byte {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write([]byte(message))
	return h.Sum(nil)
}

func makeBase64(message []byte) string {
	encoded := base64.StdEncoding.EncodeToString(message)
	return encoded
}

func stringToSign(path, query string) string {
	val := path + "\n" + query
	return val
}

func parseURI(uriString string) *url.URL {
	uri, err := url.Parse(uriString)
	if err != nil {
		panic(err)
	}
	return uri
}

func makeSignature(uri *url.URL) string {
	//Stage1: Find public Key
	public_key := getPublicKey(uri)

	//Find matching Secret Key
	secret_key := getSecretKey(public_key)

	//Construct Canonical Query
	query := canonicalQuery(uri)

	//Construct Path
	path := canonicalPath(uri)

	//Sign the strings, by joining \n
	signed_string := stringToSign(path, query)

	//Create Sha512 HMAC string
	hmac_string := makeHmac512(signed_string, secret_key)

	//Encode the resultant to base64
	base64_string := makeBase64(hmac_string)

	return base64_string
}

func IsRequestValid(url_string string) {
	url_struct := parseURI(url_string)

	isTimestampValid(url_struct)

	values := getValues(url_struct)
	sign_param, ok := values["signature"]
	values.Del("signature")
	url_struct.RawQuery = values.Encode()

	computed_signature := makeSignature(url_struct)
	log.Println("Computed: => ", computed_signature)

	if !ok || len(sign_param) < 0 {
		panic("Missing signature in request")
	}

	request_signature := sign_param[0]
	if request_signature != computed_signature {
		log.Println("Passed: => ", request_signature)
		panic("Signatures do not match.")
	}
}

func main() {
	url_string := "http://staging.moire.opschanger.nl/urls?hello=world&apple=good&access_token=%2BkfYpG%2BMQE%2BI1QAAsYyBKw&timestamp=2015-06-15T09:18:56%2B05:30&signature=SoSWX04%2BjdwJWAXdGFrI7wyahWh1WN9rumT7QrnuWpK49968mJBRh0rNmcWpU/0fnnUXUW138pM/F/QyFu9OHg=="
	IsRequestValid(url_string)
}
