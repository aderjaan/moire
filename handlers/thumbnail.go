package handlers

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/bulletind/moire/config"
	"github.com/bulletind/moire/db"

	"gopkg.in/simversity/gottp.v3"

	"os"
	"os/exec"
	"strconv"
)

type cmdStruct struct {
	Command string
	Args    []string
}

//const canvasCmd = "composite -gravity center %v %v %v"
//const iconCmd = "composite -gravity center %v %v %v"
//const mogCmd = "mogrify -resize 640x480 %v"
//
//const CANVAS_PATH = "/tmp/black_canvas.png"
//const PLAY_ICON_PATH = "/tmp/playiconhover.png"

const thumbTime int = 1
const thumbWidth int = 320
const thumbHeight int = 240

func execCommand(cmd *exec.Cmd) {
	err := executeRaw(cmd)
	if err != nil {
		panic(errors.New("Error in executable : " + err.Error()))
	}
}

//func patchPlayIcon(thumbPath string) string {
//	canvaser := fmt.Sprintf(canvasCmd, thumbPath, CANVAS_PATH, thumbPath)
//	iconer := fmt.Sprintf(iconCmd, thumbPath, PLAY_ICON_PATH, thumbPath)
//	mogrifier := fmt.Sprintf(mogCmd, thumbPath)
//
//	execCommand(canvaser)
//	execCommand(iconer)
//	execCommand(mogrifier)
//
//	return thumbPath
//}

func videoThumbnail(asset *db.Asset, duration, sizeW, sizeH int) string {
	assetId := asset.Id.Hex()

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

	signedUrl, _ := getSignedURL(asset.Bucket, asset.Path)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	cleanupThumbnail(thumbPath)

	//input := fmt.Sprintf(`%v`, signedUrl)
	time := fmt.Sprintf(`%02d:%02d:%02d`, hour, minute, second)
	scale := fmt.Sprintf(`scale=%v:-1`, sizeW)

	videoThumber := exec.Command(
		config.Settings.Moire.FFmpeg, "-i", signedUrl, "-ss",
		time, "-vframes", "1", "-vf", scale, thumbPath,
	)

	execCommand(videoThumber)

	return thumbPath
}

func getImage(rc io.Reader) (img image.Image, ft string, err error) {
	// Tries to identify an Image as type JPEG, PNG or Gif.
	// I have seen this failing so I don't know how long this is going to last.

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

func optimizeThumbnail(thumbPath string) {
	// Optimize the thumbnail to be of 70% quality.
	// It really reduces the size of the resultant image.

	optCmd := exec.Command("optipng", "-o7", "-q", thumbPath)
	executeRaw(optCmd)
}

func DownloadFile(url, loc string) error {
	out, err := os.Create(loc)
	defer out.Close()

	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}

func imageThumbnail(asset *db.Asset, width, height int) string {

	assetId := asset.Id.Hex()

	signedUrl, _ := getSignedURL(asset.Bucket, asset.Path)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)
	cleanupThumbnail(thumbPath)

	scale := fmt.Sprintf(`%vx%v`, width, height)

	if err := DownloadFile(signedUrl, thumbPath); err != nil {
		panic(err)
	}

	imageThumber := exec.Command(
		"mogrify", "-resize", scale, "-gravity", "center", "-extent", scale, "-format", "png", "-quality", "75", thumbPath,
	)

	execCommand(imageThumber)

	//rc, err := getS3Reader(bucket, path)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer rc.Close()

	//img, _, err := getImage(rc)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//m := resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)

	//out, err := os.Create(thumbPath)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//defer out.Close()

	//// write new image to file
	//if err := png.Encode(out, m); err != nil {
	//	log.Fatal(err)
	//}

	return thumbPath
}

func cleanupThumbnail(path string) error {
	return os.Remove(path)
}

const TemporaryRedirect = 302

type thumbArgs struct {
	Time string `json:"time"`
	W    string `json:"w"`
	H    string `json:"h"`
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

	valid := ValidateSignature(request)
	if valid == false {
		return
	}

	conn := getConn()
	asset, _ := getAsset(conn, _id)

	request.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	if asset.Status != db.READY && asset.FileType == ImageFile {
		pollUntilReady(conn, _id)
		asset, _ = getAsset(conn, _id)
	}

	args := thumbArgs{}
	request.ConvertArguments(&args)

	_, no_redirect := request.GetArgument("no_redirect").(string)

	thumbUrl, err := getThumbnailURL(asset)
	if err != nil {
		if no_redirect == true {
			request.Raise(gottp.HttpError{
				http.StatusNotFound,
				err.Error(),
			})

			return
		}
		// although temporary redirects should not be cached, some clients behave otherwise
		request.Writer.Header().Set("Cache-Control", "max-age=0, no-cache, no-store")
		request.Writer.Header().Set("Pragma", "no-cache")

		request.Redirect(thumbUrl, TemporaryRedirect)
		return
	}

	// If err is nil, implies that thumbnail was successfully located.
	if args.Time+args.W+args.H+args.X+args.Y != "" {
		time, _ := strconv.Atoi(args.Time)
		if time == 0 {
			time = thumbTime
		}

		w, _ := strconv.Atoi(args.W)
		x, _ := strconv.Atoi(args.X)
		// fallback to x if w is not provided
		if w == 0 && x != 0 {
			w = x
		}
		// use default width
		if w == 0 {
			w = thumbWidth
		}

		h, _ := strconv.Atoi(args.H)
		y, _ := strconv.Atoi(args.Y)
		// fallback to y if h is not provided
		if h == 0 && y != 0 {
			h = y
		}
		// use default height
		if h == 0 {
			h = thumbHeight
		}

		cacheKey := fmt.Sprintf("%v_%v_%v", time, w, h)

		if thumbUrl, ok = asset.Thumbnails[cacheKey]; !ok {

			assetId := asset.Id.Hex()
			var thumbPath string

			if asset.FileType == VideoFile {
				thumbPath = videoThumbnail(asset, time, w, h)

			} else if asset.FileType == ImageFile {
				thumbPath = imageThumbnail(asset, w, h)

			} else {
				request.Raise(gottp.HttpError{
					http.StatusConflict,
					"Can only generate thumbnails for Image and Video files.",
				})

				return
			}

			// Optimize the just genereated thumbnail for Internet.
			optimizeThumbnail(thumbPath)

			thumbUrl = path.Join("/", "thumbnail", assetId, cacheKey)
			uploadFile(asset.Bucket, thumbUrl, thumbPath)

			updateAsset(conn, assetId, db.M{"$set": db.M{"thumbnails." + cacheKey: thumbUrl}})
			cleanupThumbnail(thumbPath)
		}
	}

	signed_url, expiryDate := getSignedURL(asset.Bucket, thumbUrl)
	request.Writer.Header().Set("Expires", expiryDate.Format(time.RFC1123))
	request.Redirect(signed_url, TemporaryRedirect)
	return
}
