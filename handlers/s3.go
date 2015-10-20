package handlers

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"
	"gopkg.in/bulletind/goamz.v1/aws"
	"gopkg.in/bulletind/goamz.v1/s3"
)

func GenerateSlug(str string) (slug string) {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '.':
			return '.'
		case r == ' ', r == '-':
			return '-'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
		return -1
	}, strings.ToLower(strings.TrimSpace(str)))
}

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

func uploadFile(bucket, uploadUrl, filePath string) {
	fileType := guessMimeType(filePath)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	b := getBucket(bucket)

	log.Println("Uploading:", uploadUrl)
	b.Put(uploadUrl, data, fileType, s3.Private, s3.Options{})
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
	if strings.HasPrefix(path, "http") {
		return path
	}

	b := getBucket(bucket)
	return b.SignedURL(path, time.Now().Add(time.Hour))
}

func makeUploadURL(assetId, collection, filename, prefix string) string {
	seperator := "/"
	return path.Join(
		seperator,
		collection,
		assetId,
		prefix,
		randSeq(10),
		GenerateSlug(filename),
	)
}

func getThumbnailUploadURL(assetId, collection, fileName string) string {
	// Create a filename for the thumbnail, with PNG extension
	thumbName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	thumbFileName := fmt.Sprintf("%s.%s", thumbName, "png")

	return makeUploadURL(assetId, collection, thumbFileName, "thumbnail")
}

func getUploadURL(assetId, collection, fileName string) string {
	return makeUploadURL(assetId, collection, fileName, UploadPrefix)
}

func getThumbnailURL(asset *db.Asset) (url string, err error) {
	url = asset.ThumbnailPath
	if url == "" {
		url = thumbnailUrls(asset.FileType)
	}

	if asset.Status == db.LOST {
		url = missingUrl
		err = errors.New(thumbUnavailable)
	} else if asset.Status != db.READY {
		err = errors.New(uploadInProgress)
	}

	return
}

func getURL(asset *db.Asset) (url string, err error) {
	switch asset.Status {
	case db.READY:
		url = getSignedURL(asset.Bucket, asset.Path)
		break
	case db.LOST:
		err = errors.New(contentUnavailable)
		break
	default:
		if asset.FileType == ImageFile {
			url = thumbnailUrls(ImageFile)
		}
		err = errors.New(uploadInProgress)
		break
	}

	return
}
