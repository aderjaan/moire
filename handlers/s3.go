package handlers

import (
	"errors"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
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
	return b
}

func getSignedURL(bucket, path string) string {
	return "http://google.com"
}

func getUploadURL(asset *db.Asset) string {
	if asset.Path != "" || asset.Status == db.READY {
		panic("Asset already uploded to " + asset.Path)
	}

	var url string

	switch asset.FileType {
	case ImageFile:
		url = "/" + ImageFile + "/"
	case VideoFile:
		url = "/" + VideoFile + "/"
	case AudioFile:
		url = "/" + AudioFile + "/"
	default:
		url = "/" + PlainFile + "/"
	}

	return getSignedURL(asset.Bucket, url)
}

func getThumbnailURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		url = getSignedURL(asset.Bucket, asset.ThumbnailPath)
		break
	case db.LOST:
		err = errors.New(asset.Path + " is no longer available.")
		break
	default:
		err = errors.New("This content is still being uploaded. We appreciate your impatience")
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
		err = errors.New(asset.Path + " is no longer available.")
		break
	default:
		err = errors.New("This content is still being uploaded. We appreciate your impatience")
		break
	}

	return
}
