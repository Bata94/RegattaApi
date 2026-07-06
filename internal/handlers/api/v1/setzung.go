package api_v1

import (
	"fmt"
	"math/rand/v2"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/service"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func shuffle(array []crud.Meldung) []crud.Meldung {
	for i := range array {
		j := rand.IntN( i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func SetzungsLosung(c *handler.Context) error {
	check, err := crud.CheckMeldungSetzung(c.Request.Context(), )
	if err != nil {
		return err
	}
	if check {
		retErr := &api.BAD_REQUEST
		retErr.Msg = "Setzung bereits erledigt! Vorher reseten um zu wiederholen!"
		return retErr}

	allRennen, err := crud.GetAllRennen(c.Request.Context(), crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	for _, r := range allRennen {
		maxBahnen := 1

		switch r.Wettkampf {
		case sqlc.WettkampfKurzstrecke:
			maxBahnen = 4
		case sqlc.WettkampfSlalom:
			maxBahnen = 3
		case sqlc.WettkampfLangstrecke:
			maxBahnen = 99999
		case sqlc.WettkampfStaffel:
			maxBahnen = 2
		}

		numMeld := 0
		for _, m := range r.Meldungen {
			if !m.Abgemeldet {
				numMeld++
			}
		}
		if numMeld == 0 {
			continue
		}

		remainder := numMeld % maxBahnen
		numAbteilungen := numMeld / maxBahnen
		if remainder > 0 {
			numAbteilungen++
		}

		sizes := make([]int, numAbteilungen)
		for i := range sizes {
			sizes[i] = maxBahnen
		}
		if remainder > 0 {
			sizes[numAbteilungen-1] = remainder
		}
		if remainder == 1 && numAbteilungen >= 2 {
			sizes[numAbteilungen-2]--
			sizes[numAbteilungen-1]++
		}

		r.Meldungen = shuffle(r.Meldungen)

		abteilungIdx := 0
		bahn := int32(1)
		count := 0

		for _, m := range r.Meldungen {
			if m.Abgemeldet {
				continue
			}
			if err := crud.UpdateMeldungSetzung(c.Request.Context(), sqlc.UpdateMeldungSetzungParams{
				Uuid:      m.Uuid,
				Abteilung: int32(abteilungIdx + 1),
				Bahn:      bahn,
			}); err != nil {
				return err
			}
			bahn++
			count++
			if count >= sizes[abteilungIdx] {
				abteilungIdx++
				bahn = 1
				count = 0
			}
		}
	}

	if c.IsHtmxRequest() {
		return nil
	} else {
		return c.JSON("Setzung erfolgreich!")
	}
}

func ResetSetzung(c *handler.Context) error {
	mLs, err := crud.GetAllMeldungen(c.Request.Context(), )
	if err != nil {
		return err
	}

	for _, m := range mLs {
		err = crud.UpdateMeldungSetzung(c.Request.Context(), sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: 0,
			Bahn:      0,
		})
		if err != nil {
			return err
		}
	}

	if c.IsHtmxRequest() {
		return nil
	} else {
		return c.JSON("Losung erfolgreich!")
	}
}

func SetStartnummern(c *handler.Context) error {
	if err := service.SetStartnummern(c.Request.Context()); err != nil {
		return handler.InternalError(fmt.Sprintf( "Error while setting startnummern: %s", err.Error()))
	}

	return c.JSON("Startnummern vergeben!")
}

func SetZeitplan(c *handler.Context) error {
	param := new(service.SetZeitplanParams)
	err := c.BodyParser(param)
	if err != nil {
		return err
	}

	err = service.SetZeitplan(c.Request.Context(), *param)
	if err != nil {
		return err
	}

	return c.JSON("Zeitplan gesetzt!")
}
