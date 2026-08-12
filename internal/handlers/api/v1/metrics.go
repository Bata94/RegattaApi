package api_v1

import (
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/utils/metrics"
)

func MetricsApi(c *handler.Context) error {
	m := metrics.Collect()
	return c.JSON(m)
}
