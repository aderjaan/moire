package handlers

import (
	"errors"
	"io/ioutil"
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

func uploadFile(filePath string) {
	fileType := guessMimeType(filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	b := getBucket(config.Settings.S3.Bucket)
	b.Put(filePath, data, fileType, s3.PublicRead, s3.Options{})
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

func getSignedURL(bucket, path, mimetype string, isUpload bool) string {
	b := getBucket(bucket)
	var url = ""
	if isUpload {
		url = b.UploadSignedURL(path, "PUT", mimetype, time.Now().Add(time.Hour))
	} else {
		url = b.SignedURL(path, time.Now().Add(time.Hour))
	}
	return url
}

func getUploadURL(assetId, fileType string) string {
	var url string

	switch fileType {
	case ImageFile:
		url = ImageFile
	case VideoFile:
		url = VideoFile
	case AudioFile:
		url = AudioFile
	default:
		url = PlainFile
	}

	return path.Join(url, assetId, randSeq(10))
}

func getThumbnailURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		if asset.ThumbnailPath == "" {
			err = errors.New("Ouch! This thumbnail is no longer available.")
		}
		url = getSignedURL(asset.Bucket, asset.ThumbnailPath, asset.MimeType, false)
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
		url = getSignedURL(asset.Bucket, asset.Path, asset.MimeType, false)
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
