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
	"gopkg.in/simversity/gottp.v2/utils"

	"os"
	"os/exec"
	"strings"
)

type cmdStruct struct {
	Command string
	Args    []string
}

const thumbCmd = "ffmpeg -i \"%v\" -ss %02d:%02d:%d -vframes 1 -vf scale=-1:600 %v"

const canvasCmd = "composite -gravity center %v %v %v"
const iconCmd = "composite -gravity center %v %v %v"
const mogCmd = "mogrify -resize 640x480 %v"

const CANVAS_PATH = "/tmp/black_canvas.png"
const PLAY_ICON_PATH = "/tmp/playiconhover.png"

const thumbX uint = 320
const thumbY uint = 240

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

func videoThumbnail(assetId, bucket, url string, duration int) string {
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

	signedUrl := getSignedURL(bucket, url)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	cleanupThumbnail(thumbPath)

	videoThumber := fmt.Sprintf(thumbCmd, signedUrl, hour, minute, second, thumbPath)
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

func imageThumbnail(assetId, bucket, url string, sizeX, sizeY uint) string {
	//gcd := getGCD(sizeX, sizeY)
	//aspect := fmt.Sprintf("%v/%v", sizeX/gcd, sizeY/gcd)

	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)
	cleanupThumbnail(thumbPath)

	rc, err := getS3Reader(bucket, url)
	if err != nil {
		panic(err)
	}

	defer rc.Close()

	img, _, err := getImage(rc)
	if err != nil {
		panic(err)
	}

	m := resize.Thumbnail(sizeX, sizeY, img, resize.Lanczos3)

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
	Time      int  `json:"time" required:"true"`
	PatchIcon bool `json:"patch_icon" required:"thumb_time"`
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
	//Allows you to create a new thumbnial for the Video files at the provided
	//time (in seconds) argument.
	//This POST call is pretty much meaningless for all non-video file formats.

	args := thumbArgs{}
	request.ConvertArguments(&args)

	errs := utils.Validate(&args)
	if len(*errs) > 0 {
		request.Raise(gottp.HttpError{
			http.StatusBadRequest,
			ConcatenateErrors(errs),
		})

		return
	}

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
		thumbPath = videoThumbnail(assetId, asset.Bucket, asset.Path, args.Time)

	} else if asset.FileType == ImageFile {
		thumbPath = imageThumbnail(assetId, asset.Bucket, asset.Path, thumbX, thumbY)

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
