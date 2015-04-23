package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"gopkg.in/simversity/gottp.v2"

	"os/exec"
	"strings"
)

type cmdStruct struct {
	Command string
	Args    []string
}

const thumbCmd = "ffmpeg -i %v -ss 00:00:01 -vframes 1 -vf scale=-1:600 %v"
const canvasCmd = "composite -gravity center %v %v %v"
const iconCmd = "composite -gravity center %v %v %v"
const mogCmd = "mogrify -resize 640x480 %v"

const CANVAS_PATH = "/tmp/black_canvas.png"
const PLAY_ICON_PATH = "/tmp/playiconhover.png"

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

func videoThumbnail(bucket, url string) string {
	thumbPath := "/tmp/thumbPath.png"
	signedUrl := getSignedURL(bucket, url)

	videoThumber := fmt.Sprintf(thumbCmd, signedUrl, thumbPath)
	execCommand(videoThumber)

	return thumbPath
}

func imageThumbnail(bucket, url string) string {
	thumbPath := "/tmp/thumbPath.png"
	signedUrl := getSignedURL(bucket, url)

	imageThumber := fmt.Sprintf(thumbCmd, signedUrl, thumbPath)
	execCommand(imageThumber)

	return thumbPath
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
