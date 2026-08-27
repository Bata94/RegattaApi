package components

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func ImageComponent(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	src := queryParams.Get("src")
	alt := queryParams.Get("alt")

	if src == "" {
		webfw.ErrorToast(w, r, "Image src is empty")
		return
	}

	imgOpt := ui_components.DefaultImageOptions()

	if w := queryParams.Get("width"); w != "" {
		imgOpt.Width = w
	}
	if h := queryParams.Get("height"); h != "" {
		imgOpt.Height = h
	}
	if q := queryParams.Get("quality"); q != "" {
		qFloat64, err := strconv.ParseFloat(q, 32)
		if err != nil || qFloat64 <= 0.0 || qFloat64 > 100.0 {
			slog.Warn("Image component: quality is not a float32 value or out of range... Setting default value")
		} else {
			imgOpt.Quality = float32(qFloat64)
		}
	}
	if l := queryParams.Get("lossless"); l != "" {
		imgOpt.Lossless, _ = strconv.ParseBool(l)
		if imgOpt.Lossless {
			slog.Warn("Image component: lossless is not a bool value... Setting default value")
			imgOpt.Lossless = false
		}
	}
	if class := queryParams.Get("class"); class != "" {
		imgOpt.ClassImage = class
	}

	webfw.SetCacheControl(w, "public, max-age=31536000, immutable")
	templ.Handler(ui_components.RawImageComponent(src, alt, imgOpt)).ServeHTTP(w, r)
}
