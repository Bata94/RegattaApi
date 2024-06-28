package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

func DrvMeldungUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		log.Error(err.Error())
		return &api.BAD_REQUEST
	}

	log.Info(file.Filename)
	uploadsDir := "./tmp/uploads/"

	err = os.MkdirAll(uploadsDir, 0o666)
	if err != nil {
		return err
	}

	dest := fmt.Sprintf("%s%s_%s.json", uploadsDir, "DrvMeldung", time.Now().Format("2006-01-02_15-04-05"))
	err = c.SaveFile(file, dest)
	if err != nil {
		log.Error(err.Error())
		return &api.INTERNAL_SERVER_ERROR
	}

	err = ImportDrvJson(dest)
	if err != nil {
		log.Error("TopLevelError!")
		return &api.ReqError{
			Code:       500,
			StatusCode: fiber.StatusInternalServerError,
			Title:      api.INTERNAL_SERVER_ERROR.Title,
			Msg:        "An Error accurred while importing the JSON File! If you directly downloaded the File from DRV and uploaded it, without modifying it, please contact the Admin!",
			Details:    err.Error(),
			Data:       nil,
		}
	}

	return c.JSON("File uploaded successfully!")
}

type DrvMeldungJson struct {
	Metadata    DrvMetadata      `json:"_metadata"`
	Regatta     DrvRegatta       `json:"regatta"`
	Entries     []DrvEntries     `json:"entries"`
	Events      []DrvEvents      `json:"events"`
	Clubs       []DrvClubs       `json:"clubs"`
	ClubMembers []DrvClubMembers `json:"club_members"`
	ClubBoats   []DrvClubBoats   `json:"club_boats"`
}

type DrvMetadata struct {
	TimeCreated   time.Time `json:"timestamp"`
	FormatVersion string    `json:"format_version"`
}

type DrvRegatta struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	City     string    `json:"city"`
	Website  string    `json:"website"`
	Language string    `json:"language"`
	Days     []string  `json:"days"`
}

type DrvEntries struct {
	Id         uuid.UUID `json:"id"`
	RevisionId uuid.UUID `json:"revision_id"`
	EventId    uuid.UUID `json:"event_id"`
	ClubId     uuid.UUID `json:"club_id"`
	Name       string    `json:"name"`
	ShortName  string    `json:"shortname"`
	Sequence   int       `json:"sequence"`
	// combination, alternative, status, alternative_event_id ???
	Members []DrvEntriesMembers `json:"members"`
}

type DrvEntriesMembers struct {
	ClubMemberId uuid.UUID `json:"club_member_id"`
	Role         string    `json:"role"`
	Position     int       `json:"position"`
}

type DrvEvents struct {
	Id       uuid.UUID        `json:"id"`
	Number   string           `json:"number"`
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Addition string           `json:"addition"`
	Sex      string           `json:"sex"`
	Weighed  bool             `json:"weighed"`
	Days     []DrvEventDay    `json:"days"`
	Remarks  string           `json:"remarks"`
	Category DrvEventCategory `json:"category"`
	BoatType DrvEventBoatType `json:"boattype"`
	Cost     DrvEventCost     `json:"cost"`
}

type DrvEventCategory struct {
	Id   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type DrvEventBoatType struct {
	Id   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}
type DrvEventCost struct {
	Id       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Currency string    `json:"currency"`
	Amount   int       `json:"amount"`
}

type DrvEventDay struct {
	Date      string `json:"day_date"`
	SortOrder int    `json:"sort_order"`
}

type DrvClubs struct {
	Id        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
}

type DrvClubMembers struct {
	Id     uuid.UUID `json:"id"`
	ClubId uuid.UUID `json:"club_id"`
	Person DrvPerson `json:"person"`
}

type DrvPerson struct {
	Id          uuid.UUID `json:"id"`
	Firstname   string    `json:"firstname"`
	Lastname    string    `json:"lastname"`
	Sex         string    `json:"sex"`
	YearOfBirth string    `json:"yearofbirth"`
}

type DrvClubBoats struct{}

func ImportDrvJson(filePath string) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Error("Read File Error")
		log.Error(err.Error())
		return err
	}

	drvMeldung := DrvMeldungJson{}
	err = json.Unmarshal(b, &drvMeldung)
	if err != nil {
		log.Error("Unmarshal Error")
		log.Error(err.Error())
		return err
	}

	o, err := json.MarshalIndent(drvMeldung, "", "  ")
	if err != nil {
		log.Error("Marshal Error")
		log.Error(err.Error())
		return err
	}
	os.WriteFile(fmt.Sprintf("./tmp/Imported_DrvMeldung_%s.json", time.Now().Format("15-04-05")), o, 0o666)

	var apiReqError *api.ReqError
	// TODO: Use a Transaction here!

	for _, v := range drvMeldung.Clubs {
		verein, err := crud.GetVereinMinimal(v.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
			}
		}
		if verein != nil {
			log.Debug("Verein already exists: ", verein.Name)
			continue
		}

		newVerein := sqlc.CreateVereinParams{
			Uuid:     v.Id,
			Name:     &v.Name,
			Kurzform: &v.ShortName,
			Kuerzel:  &v.Code,
		}
		_, err = crud.CreateVerein(newVerein)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	for _, r := range drvMeldung.Events {
		rennen, err := crud.GetRennenMinimal(r.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
			}
		}
		if rennen != nil {
			log.Debugf("Rennen already exists: %s - %s", rennen.Nummer, *rennen.BezeichnungLang)
			continue
		}

		var rennabstand, kosten int32
		var sex string

		// TODO: Set Rennabstand, Tag, Wettkampf properly
		rennabstand = 5
		kosten = int32(r.Cost.Amount)

		if r.Sex == "" {
			sex = "x"
		} else {
			sex = strings.ToLower(r.Sex)
		}

		newRennen := sqlc.CreateRennenParams{
			Uuid:            r.Id,
			SortID:          int32(r.Days[0].SortOrder),
			Nummer:          r.Number,
			Bezeichnung:     &r.Code,
			BezeichnungLang: &r.Name,
			Zusatz:          &r.Addition,
			Leichtgewicht:   &r.Weighed,
			Geschlecht: sqlc.NullGeschlecht{
				Geschlecht: sqlc.Geschlecht(sex),
				Valid:      true,
			},
			Bootsklasse:      &r.BoatType.Code,
			BootsklasseLang:  &r.BoatType.Name,
			Altersklasse:     &r.Category.Code,
			AltersklasseLang: &r.Category.Name,
			Tag:              sqlc.NullTag{},
			Wettkampf:        sqlc.NullWettkampf{},
			KostenEur:        &kosten,
			Rennabstand:      &rennabstand,
		}

		_, err = crud.CreateRennen(newRennen)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	for _, a := range drvMeldung.ClubMembers {
		athlet, err := crud.GetAthletMinimal(a.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
			}
		}
		if athlet != nil {
			log.Debugf("Athlet already exists: %s %s", athlet.Vorname, athlet.Name)
			continue
		}

		var arztBesch bool
		arztBesch = true

		newAthlet := sqlc.CreateAthletParams{
			Uuid:                    a.Id,
			VereinUuid:              a.ClubId,
			Name:                    a.Person.Firstname,
			Vorname:                 a.Person.Lastname,
			Jahrgang:                a.Person.YearOfBirth,
			AerztlicheBescheinigung: &arztBesch,
			Geschlecht:              sqlc.Geschlecht(strings.ToLower(a.Person.Sex)),
		}

		_, err = crud.CreateAthlet(newAthlet)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	for _, m := range drvMeldung.Entries {
		meldung, err := crud.GetMeldungMinimal(m.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
			}
		}
		if meldung != nil {
			log.Debug("Meldung already exists...")
			// TODO: Compare RevisionID to easy import Changes from Meldeportal
			continue
		}

		// TODO: Acount for alternative Meldung
		typ := "DRV Meldung"
		bemerkung := ""
    athleten := []crud.MeldungAthlet{}

    for _, a := range m.Members {
      role := strings.ToLower(a.Role)
      position := int32(a.Position)
      if role == "cox" {
        athleten = append(athleten, crud.MeldungAthlet{
        	Uuid:     a.ClubMemberId,
        	Position: &position,
        })
      } else if role == "coach" {
        log.Debug("Role Coach not implemented")
        continue
      } else if role == "rower" {
        athleten = append(athleten, crud.MeldungAthlet{
        	Uuid:     a.ClubMemberId,
        	Position: &position,
        })
      } else {
        log.Error("Unkown Role: ", a.Role)
      }
    }

		newMeldung := crud.CreateMeldungParams{
			CreateMeldungParams: &sqlc.CreateMeldungParams{
        Uuid:            m.Id,
        VereinUuid:      m.ClubId,
        RennenUuid:      m.EventId,
        DrvRevisionUuid: m.RevisionId,
        Typ:             &typ,
        Bemerkung:       &bemerkung,
      },
			Athleten:          athleten,
		}

		_, err = crud.CreateMeldung(newMeldung)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	return nil
}
