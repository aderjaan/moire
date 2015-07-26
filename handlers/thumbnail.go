package handlers

import (
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"net/http"
	"path"

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
const thumbX int = 320
const thumbY int = 240

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

func videoThumbnail(asset *db.Asset, duration, sizeX, sizeY int) string {
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

	signedUrl := getSignedURL(asset.Bucket, asset.Path)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)

	cleanupThumbnail(thumbPath)

	//input := fmt.Sprintf(`%v`, signedUrl)
	time := fmt.Sprintf(`%02d:%02d:%02d`, hour, minute, second)
	scale := fmt.Sprintf(`scale=%v:-1`, sizeX)

	videoThumber := exec.Command(
		config.Settings.Moire.FFmpeg, "-i", signedUrl, "-ss",
		time, "-vframes", "1", "-vf", scale, thumbPath,
	)

	execCommand(videoThumber)

	return thumbPath
}

func getGCD(x, y int64) int64 {
	// I have no idea, where this function is being used.

	gcd := new(big.Int).GCD(nil, nil, big.NewInt(x), big.NewInt(y)).Int64()
	return gcd
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
	out, err := os.Create("output.txt")
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

func imageThumbnail(asset *db.Asset, sizeX, sizeY int) string {

	assetId := asset.Id.Hex()

	signedUrl := getSignedURL(asset.Bucket, asset.Path)
	thumbPath := fmt.Sprintf("/tmp/%v_thumb.png", assetId)
	cleanupThumbnail(thumbPath)

	scale := fmt.Sprintf(`scale='if(gt(a,4/3),%v,-1)':'if(gt(a,4/3),-1,%v)'`, sizeX, sizeY)

	var loc string

	if asset.FileType == ImageFile {
		loc = path.Join("/", "tmp", assetId+"_"+randSeq(10))
		err := DownloadFile(signedUrl, loc)
		defer cleanupThumbnail(loc)

		if err != nil {
			panic(err)
		}

	} else {
		loc = signedUrl
	}

	imageThumber := exec.Command(
		config.Settings.Moire.FFmpeg, "-i", loc, "-vf", scale, thumbPath,
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

	//m := resize.Thumbnail(uint(sizeX), uint(sizeY), img, resize.Lanczos3)

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

	if asset.Status != db.READY && asset.FileType == ImageFile {
		pollUntilReady(conn, _id)
		asset = getAsset(conn, _id)
	}

	args := thumbArgs{}
	request.ConvertArguments(&args)

	valid := ValidateSignature(request)
	if valid == false {
		return
	}

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

		// If err is nil, implies that thumbnail was successfully located.
		request.Redirect(thumbUrl, TemporaryRedirect)
		return
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

			// Optimize the just genereated thumbnail for Internet.
			optimizeThumbnail(thumbPath)

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
