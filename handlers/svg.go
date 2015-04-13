package handlers

import (
	"net/http"

	"github.com/ajstarks/svgo"
)

func getPlaceHolder(w http.ResponseWriter, message string) {
	width := 640
	height := 480

	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(width, height)
	s.CenterRect(width/2, height/2, width, height, "fill:none;stroke:black")
	s.Text(width/2, height/2, message, "text-anchor:middle;")
	s.End()
}
