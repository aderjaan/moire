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

func execCommand(command []string) {
	err := exec.Command(command[0], command[1:len(command)]...).Run()
	if err != nil {
		panic(errors.New("Error in executable : " + command[0] + " " + err.Error()))
	}
}

func patchPlayIcon(thumbPath string) string {
	canvaser := strings.Split(fmt.Sprintf(canvasCmd, thumbPath, CANVAS_PATH, thumbPath), " ")
	iconer := strings.Split(fmt.Sprintf(iconCmd, thumbPath, PLAY_ICON_PATH, thumbPath), " ")
	mogrifier := strings.Split(fmt.Sprintf(mogCmd, thumbPath), " ")
	log.Println(canvaser, iconer, mogrifier)

	execCommand(canvaser)
	execCommand(iconer)
	execCommand(mogrifier)

	return thumbPath
}

func generateThumbnail(bucket, videoUrl string) string {
	thumbPath := "/tmp/thumbPath.png"
	url := getSignedURL(bucket, videoUrl)
	thumber := strings.Split(fmt.Sprintf(thumbCmd, url, thumbPath), " ")

	log.Println(thumber)

	execCommand(thumber)

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
