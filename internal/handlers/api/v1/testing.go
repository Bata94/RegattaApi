package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/mailer"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	err := mailer.Enqueue(r.Context(), mailer.Params{
		To: []string{
			"bastian.sievers@gmail.com",
			"bastian.sievers+test@gmail.com",
		},
		CC:      []string{},
		Subject: "Test Mail",
		Body:    "Dies ist eine TestMail",
		Files:   []string{},
	})

	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "success")
}
