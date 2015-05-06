package handlers

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"path"
	"path/filepath"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
)

const UploadPrefix = "original_file"

func getRegion() aws.Region {
	return aws.Regions[config.Settings.S3.Region]
}

func getBucket(bucket string) *s3.Bucket {
	auth := aws.Auth{
		AccessKey: config.Settings.S3.AccessKey,
		SecretKey: config.Settings.S3.SecretKey,
	}

	region := getRegion()

	connection := s3.New(auth, region)
	b := connection.Bucket(bucket)
	if b == nil {
		panic("Bucket not found.")
	}

	return b
}

func guessMimeType(filePath string) string {
	ext := filepath.Ext(filePath)
	return mime.TypeByExtension(ext)
}

func uploadFile(uploadUrl, filePath string) string {
	fileType := guessMimeType(filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	b := getBucket(config.Settings.S3.Bucket)

	log.Println("Uploaading:", uploadUrl)
	b.Put(uploadUrl, data, fileType, s3.PublicRead, s3.Options{})

	return getSignedURL(config.Settings.S3.Bucket, uploadUrl)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getSignedUploadURL(bucket, path, mimetype string) string {
	expiry := time.Hour * 24 * 365
	b := getBucket(bucket)
	url := b.UploadSignedURL(path, "PUT", mimetype, time.Now().Add(expiry))
	return url
}

func getS3Reader(bucket, path string) (io.ReadCloser, error) {
	b := getBucket(bucket)
	return b.GetReader(path)
}

func getSignedURL(bucket, path string) string {
	b := getBucket(bucket)
	return b.SignedURL(path, time.Now().Add(time.Hour))
}

func getUploadURL(assetId string) string {
	seperator := "/"
	return path.Join(seperator, UploadPrefix, assetId, randSeq(10))
}

func getThumbnailURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		if asset.ThumbnailPath == "" {
			err = errors.New("Ouch! This thumbnail is no longer available.")
		} else {
			url = getSignedURL(asset.Bucket, asset.ThumbnailPath)
		}
		break
	case db.LOST:
		err = errors.New("Ouch! This thumbnail is no longer available.")
		break
	default:
		err = errors.New("Hmm! Thumbnail was not found as the content is still being uploaded.")
		break
	}

	return
}

func getURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		url = getSignedURL(asset.Bucket, asset.Path)
		break
	case db.LOST:
		err = errors.New("Ouch! This content is no longer available.")
		break
	default:
		err = errors.New("This content is still being uploaded. We appreciate your impatience")
		break
	}

	return
}
