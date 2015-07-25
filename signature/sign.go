package signature

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"github.com/bulletind/moire/logger"
	"net/url"
	"strings"
	"time"
)

var log = logger.Logger

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

	log.Debug("Timestamp valid?", "current", current_time, "Timestamp:", timestamp, "Skew", max_time_skew, "Offset", max_time_offset)

	if timestamp.Sub(max_time_skew) > 0 {
		return errors.New("Timestamp max skew validation error")
	}

	if timestamp.Sub(max_time_offset) < 0 {
		return errors.New("Timestamp max offset validation error")
	}

	return nil
}

func canonicalQuery(public_key, timestamp string) string {
	values := url.Values{"public_key": {public_key}, "timestamp": {timestamp}}
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

func MakeSignature(public_key, secret_key, timestamp, path string) string {
	//Stage1: Find public Key

	//Construct Canonical Query
	query := canonicalQuery(public_key, timestamp)
	log.Debug("Generating Signature", "Query", query, "Path", path)

	//Sign the strings, by joining \n
	signed_string := stringToSign(path, query)
	log.Debug("Generating Signature", "SignedString", signed_string)

	//Create Sha512 HMAC string
	hmac_string := makeHmac512(signed_string, secret_key)
	log.Debug("Generating Signature", "Hmac string", hmac_string)

	//Encode the resultant to base64
	base64_string := makeBase64(hmac_string)

	return base64_string
}

func IsRequestValid(
	public_key, private_key, timestamp, signature, path string,
) error {

	err := isTimestampValid(timestamp)
	if err != nil {
		return err
	}

	computed_signature := MakeSignature(public_key, private_key, timestamp, path)

	log.Debug("IsRequestValid", "PublicKey", public_key, "PrivateKey", private_key, "Computed", computed_signature)

	if signature != computed_signature {
		return errors.New("Invalid signature")
	}

	return nil
}
