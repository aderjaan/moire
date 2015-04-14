package handlers

import (
	"errors"
	"math/rand"
	"time"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"gopkg.in/amz.v3/aws"
	"gopkg.in/amz.v3/s3"
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
	b, err := connection.Bucket(bucket)
	if err != nil {
		panic(err)
	}

	return b
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getSignedURL(bucket, path string) string {
	b := getBucket(bucket)
	url, err := b.SignedURL(path, 60*time.Minute)
	if err != nil {
		panic(err)
	}

	return url
}

func getUploadURL(asset *db.Asset) string {
	if asset.Path != "" || asset.Status == db.READY {
		panic("Asset already uploded to " + asset.Path)
	}

	var url string

	switch asset.FileType {
	case ImageFile:
		url = ImageFile
	case VideoFile:
		url = VideoFile
	case AudioFile:
		url = AudioFile
	default:
		url = PlainFile
	}

	url += "/" + asset.Id.Hex() + "/" + randSeq(10)

	return getSignedURL(asset.Bucket, url)
}

func getThumbnailURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		url = getSignedURL(asset.Bucket, asset.ThumbnailPath)
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