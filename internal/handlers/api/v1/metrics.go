package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/utils/metrics"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func MetricsApi(w http.ResponseWriter, r *http.Request) {
	m := metrics.Collect()
	webfw.JSON(w, r, m)
}
