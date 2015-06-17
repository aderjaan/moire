package signature

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"
)

func getValues(uri *url.URL) url.Values {
	return uri.Query()
}

func isTimestampValid(signed_on string) error {
	timestamp, err := time.Parse(time.RFC3339, signed_on)
	if err != nil {
		return err
	}

	current_time := time.Now()

	max_time_skew := current_time.Add(5 * time.Minute)
	max_time_offset := current_time.Add(-60 * time.Minute)

	log.Println("Current:", current_time)
	log.Println("Timestamp:", timestamp)
	log.Println("Skew", max_time_skew)
	log.Println("Offset", max_time_offset)

	if timestamp.Sub(max_time_skew) > 0 {
		return errors.New("Timestamp max skew validation error")
	}

	if timestamp.Sub(max_time_offset) < 0 {
		return errors.New("Timestamp max offset validation error")
	}

	return nil
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

func makeSignature(public_key, secret_key string, uri *url.URL) string {
	//Stage1: Find public Key

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

func IsRequestValid(
	public_key, private_key, timestamp, signature string,
	url_struct *url.URL,
) error {

	err := isTimestampValid(timestamp)
	if err != nil {
		return err
	}

	values := getValues(url_struct)
	values.Del("signature")
	url_struct.RawQuery = values.Encode()

	computed_signature := makeSignature(public_key, private_key, url_struct)
	log.Println("Computed: => ", computed_signature)

	if signature != computed_signature {
		log.Println("Passed: => ", signature)
		return errors.New("Invalid signature")
	}

	return nil
}
