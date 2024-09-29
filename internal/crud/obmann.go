package crud

import (
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type Obmann struct {
	*sqlc.Obmann
}

func GetAllObmannForVerein(vereinUuid uuid.UUID) ([]Obmann, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	retLs := []Obmann{}
	q, err := DB.Queries.GetAllObmannForVerein(ctx, vereinUuid)
	if err != nil {
		return retLs, err
	}

	for _, o := range q {
		retLs = append(retLs, Obmann{&o})
	}
	return retLs, nil
}
