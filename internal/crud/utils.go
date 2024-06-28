package crud

import (
	"context"
	"strings"
	"time"
)

func isNoRowError(err error) bool {
  return strings.Contains(err.Error(), "no rows in result set")
}

func getCtxWithTo() (context.Context, context.CancelFunc) {
  return context.WithTimeout(context.Background(), 60*time.Second)
}
