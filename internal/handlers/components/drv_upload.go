package components

import (
	"net/http"

	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func DrvUploadPost(w http.ResponseWriter, r *http.Request) {
	api_v1.DrvMeldungUpload(w, r)
	webfw.SuccessToast(w, r, "Upload erfolgreich!")
}
