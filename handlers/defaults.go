package handlers

import (
	"github.com/bulletind/moire/config"
	"path"
)

const (
	AudioFile = "audio"
	ImageFile = "image"
	VideoFile = "video"
	PlainFile = "plain"
)

const UploadPrefix = "original_file"

func thumbnailUrls(fileType string) string {
	switch fileType {
	case AudioFile:
		return path.Join(config.Settings.Moire.StaticPath, "/placeholders/audio.png")
	case VideoFile:
		return path.Join(config.Settings.Moire.StaticPath, "/placeholders/video.png")
	case ImageFile:
		return path.Join(config.Settings.Moire.StaticPath, "/placeholders/thumbnail.png")
	default:
		return path.Join(config.Settings.Moire.StaticPath, "/placeholders/attachment.png")
	}
}

const missingUrl = "https://s3-eu-west-1.amazonaws.com/sc-gallery/missing.png"

const thumbUnavailable = "Ouch! This thumbnail is no longer available."
const uploadInProgress = "This content is still being uploaded. We appreciate your impatience"
const contentUnavailable = "Ouch! This content is no longer available."
