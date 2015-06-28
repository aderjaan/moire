package handlers

const (
	AudioFile = "audio"
	ImageFile = "image"
	VideoFile = "video"
	PlainFile = "plain"
)

const UploadPrefix = "original_file"

var thumbnailUrls = map[string]string{
	AudioFile: "http://s3-eu-west-1.amazonaws.com/sc-gallery/attachment.png",
	VideoFile: "http://s3-eu-west-1.amazonaws.com/sc-gallery/thumbnail.png",
	ImageFile: "http://s3-eu-west-1.amazonaws.com/sc-gallery/thumbnail.png",
	PlainFile: "http://s3-eu-west-1.amazonaws.com/sc-gallery/attachment.png",
}

const missingUrl = "http://s3-eu-west-1.amazonaws.com/sc-gallery/missing.png"

const thumbUnavailable = "Ouch! This thumbnail is no longer available."
const uploadInProgress = "This content is still being uploaded. We appreciate your impatience"
const contentUnavailable = "Ouch! This content is no longer available."
