package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/bulletind/moire/db"
	"gopkg.in/simversity/gottp.v2"

	"os"
	"os/exec"
	"strings"
)

type cmdStruct struct {
	Command string
	Args    []string
}

const thumbCmd = "ffmpeg -i %v -ss 00:00:1 -vframes 1 -vf scale=-1:600 %v"
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

func videoThumbnail(assetId, bucket, url string) string {
	signedUrl := getSignedURL(bucket, url)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	videoThumber := fmt.Sprintf(thumbCmd, signedUrl, thumbPath)
	execCommand(videoThumber)

	return thumbPath
}

func imageThumbnail(assetId, bucket, url string) string {
	signedUrl := getSignedURL(bucket, url)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	imageThumber := fmt.Sprintf(thumbCmd, signedUrl, thumbPath)
	execCommand(imageThumber)

	return thumbPath
}

func cleanupThumbnail(path string) error {
	return os.Remove(path)
}

const TemporaryRedirect = 302

type thumbArgs struct {
	ThumbTime     string `json:"thumb_time" required:"true"`
	PatchPlayIcon bool   `json:"patch_play_icon" required:"thumb_time"`
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

	if asset.FileType != VideoFile && asset.FileType != ImageFile {
		request.Raise(gottp.HttpError{
			http.StatusNotFound,
			"Can only generate thumbnails for Image and Video files.",
		})

		return
	}

	url, err := getThumbnailURL(asset)
	if err != nil {
		getPlaceHolder(request.Writer, err.Error())
	} else {
		request.Redirect(url, TemporaryRedirect)
	}

	return
}

func (self *Thumbnail) Post(request *gottp.Request) {
	args := thumbArgs{}
	request.ConvertArguments(&args)

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

	assetId := asset.Id.Hex()

	var thumbPath string

	if asset.FileType == VideoFile {
		thumbPath = videoThumbnail(assetId, asset.Bucket, asset.Path)
	} else if asset.FileType == ImageFile {
		thumbPath = imageThumbnail(assetId, asset.Bucket, asset.Path)
	} else {
		request.Raise(gottp.HttpError{
			http.StatusConflict,
			"Can only generate thumbnails for Image and Video files.",
		})

		return
	}

	_, err := getThumbnailURL(asset)
	if err != nil {
		getPlaceHolder(request.Writer, err.Error())
		return
	}

	uploadUrl := path.Join("/", "thumbnail", assetId)
	signed_url := uploadFile(uploadUrl, thumbPath)
	updateAsset(conn, assetId, db.M{"$set": db.M{"thumbnail_path": uploadUrl}})
	cleanupThumbnail(thumbPath)

	request.Write(signed_url)
	//request.Redirect(signed_url, TemporaryRedirect)

	return
}
