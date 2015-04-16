package handlers

import (
	"fmt"
	"log"
	"net/http"

	"gopkg.in/simversity/gottp.v2"
)

const thumbCmd = "ffmpeg -i %v -ss 00:00:01 -vframes 1 -vf scale=-1:600 %v"
const canvasCmd = "composite -gravity center %v %v %v"
const iconCmd = "composite -gravity center %v %v %v"
const mogCmd = "mogrify -resize 640x480 %v"

const CANVAS_PATH = "/tmp/black_canvas.png"
const PLAY_ICON_PATH = "/tmp/playiconhover.png"

func patchPlayIcon(thumbPath string) {
	canvaser := fmt.Sprintf(canvasCmd, thumbPath, CANVAS_PATH, thumbPath)
	mogrifier := fmt.Sprintf(mogCmd, thumbPath)
	log.Println(thumber, canvaser, mogrifier)
}

func generateThumbnail(bucket, videoUrl string) {
	thumbPath := "/tmp/thumbPath.png"
	url := getSignedURL(bucket, videoUrl)

	thumber := fmt.Sprintf(thumbCmd, url, thumbPath)

	uploadFile(thumbPath)
}

type Thumbnail struct {
	gottp.BaseHandler
}

func (self *Thumbnail) Get(request *gottp.Request) {
	_id, ok := request.GetArgument("_id").(string)
	if !ok {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			"Not found",
		})

		return
	}

	conn := getConn()
	asset := getAsset(conn, _id)

	url, err := getThumbnailURL(asset)
	if err != nil {
		getPlaceHolder(request.Writer, err.Error())
	} else {
		request.Redirect(url, 302)
	}

	return
}
