package handlers

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/big"
	"net/http"
	"path"

	"github.com/bulletind/moire/db"
	"github.com/nfnt/resize"
	"gopkg.in/simversity/gottp.v2"

	"os"
	"os/exec"
	"strconv"
	"strings"
)

type cmdStruct struct {
	Command string
	Args    []string
}

const thumbCmd = "ffmpeg -i \"%v\" -ss %02d:%02d:%d -vframes 1 -vf scale=\"%v:-1\" %v"

const canvasCmd = "composite -gravity center %v %v %v"
const iconCmd = "composite -gravity center %v %v %v"
const mogCmd = "mogrify -resize 640x480 %v"

const CANVAS_PATH = "/tmp/black_canvas.png"
const PLAY_ICON_PATH = "/tmp/playiconhover.png"

const thumbTime int = 1
const thumbX int = 320
const thumbY int = 240

func execCommand(command string) {
	log.Println("Executing:", command)
	cmd := strings.Split(command, " ")
	err := exec.Command(cmd[0], cmd[1:len(cmd)]...).Run()
	if err != nil {
		panic(errors.New("Error in executable : " + cmd[0] + " " + err.Error()))
	}
}

func patchPlayIcon(thumbPath string) string {
	canvaser := fmt.Sprintf(canvasCmd, thumbPath, CANVAS_PATH, thumbPath)
	iconer := fmt.Sprintf(iconCmd, thumbPath, PLAY_ICON_PATH, thumbPath)
	mogrifier := fmt.Sprintf(mogCmd, thumbPath)

	execCommand(canvaser)
	execCommand(iconer)
	execCommand(mogrifier)

	return thumbPath
}

func videoThumbnail(asset *db.Asset, duration, sizeX, sizeY int) string {
	assetId := asset.Id
	bucket := asset.Bucket
	path := asset.Path

	hour := 0
	minute := 0
	second := duration

	if second > 60 {
		minute = duration / 60
		second = duration % 60
	}

	if minute > 60 {
		hour = minute / 60
		minute = minute % 60
	}

	signedUrl := getSignedURL(bucket, path)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	cleanupThumbnail(thumbPath)

	videoThumber := fmt.Sprintf(thumbCmd, signedUrl, hour, minute, second, sizeX, thumbPath)
	execCommand(videoThumber)

	return thumbPath
}

func getGCD(x, y int64) int64 {
	gcd := new(big.Int).GCD(nil, nil, big.NewInt(x), big.NewInt(y)).Int64()
	return gcd
}

func getImage(rc io.Reader) (img image.Image, ft string, err error) {
	img, ft, err = image.Decode(rc)
	if err == nil {
		return
	}

	img, err = png.Decode(rc)
	if err == nil {
		ft = "png"
		return
	}

	img, err = jpeg.Decode(rc)
	if err == nil {
		ft = "jpeg"
		return
	}

	img, err = gif.Decode(rc)
	if err == nil {
		ft = "gif"
	}

	return
}

func imageThumbnail(asset *db.Asset, sizeX, sizeY int) string {
	assetId := asset.Id
	bucket := asset.Bucket
	path := asset.Path

	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)
	cleanupThumbnail(thumbPath)

	rc, err := getS3Reader(bucket, path)
	if err != nil {
		log.Fatal(err)
	}

	defer rc.Close()

	img, _, err := getImage(rc)
	if err != nil {
		log.Fatal(err)
	}

	m := resize.Thumbnail(uint(sizeX), uint(sizeY), img, resize.Lanczos3)

	out, err := os.Create(thumbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	// write new image to file
	if err := png.Encode(out, m); err != nil {
		log.Fatal(err)
	}

	return thumbPath
}

func cleanupThumbnail(path string) error {
	return os.Remove(path)
}

const TemporaryRedirect = 302

type thumbArgs struct {
	Time string `json:"time"`
	X    string `json:"x"`
	Y    string `json:"y"`
}

type Thumbnail struct {
	gottp.BaseHandler
}

func (self *Thumbnail) Get(request *gottp.Request) {
	_id, ok := request.GetArgument("_id").(string)
	if !ok {
		request.Raise(gottp.HttpError{http.StatusNotFound, "Not found"})
		return
	}

	conn := getConn()
	asset := getAsset(conn, _id)

	args := thumbArgs{}
	request.ConvertArguments(&args)

	var thumbUrl string

	signedUrl, err := getThumbnailURL(asset)
	if err != nil {
		request.Redirect(signedUrl, TemporaryRedirect)
		return
	} else if signedUrl != "" {
		thumbUrl = signedUrl
	}

	if args.Time+args.X+args.Y != "" {
		time, _ := strconv.Atoi(args.Time)
		if time == 0 {
			time = thumbTime
		}

		x, _ := strconv.Atoi(args.X)
		if x == 0 {
			x = thumbX
		}

		y, _ := strconv.Atoi(args.Y)
		if y == 0 {
			y = thumbY
		}

		cacheKey := fmt.Sprintf("%v_%v_%v", time, x, y)

		if thumbUrl, ok = asset.Thumbnails[cacheKey]; !ok {

			assetId := asset.Id.Hex()
			var thumbPath string

			if asset.FileType == VideoFile {
				thumbPath = videoThumbnail(asset, time, x, y)

			} else if asset.FileType == ImageFile {
				thumbPath = imageThumbnail(asset, x, y)

			} else {
				request.Raise(gottp.HttpError{
					http.StatusConflict,
					"Can only generate thumbnails for Image and Video files.",
				})

				return
			}

			thumbUrl = path.Join("/", "thumbnail", assetId, cacheKey)
			uploadFile(asset.Bucket, thumbUrl, thumbPath)

			updateAsset(conn, assetId, db.M{"$set": db.M{"thumbnails." + cacheKey: thumbUrl}})
			cleanupThumbnail(thumbPath)
		}
	}

	signed_url := getSignedURL(asset.Bucket, thumbUrl)
	request.Redirect(signed_url, TemporaryRedirect)
	return
}
