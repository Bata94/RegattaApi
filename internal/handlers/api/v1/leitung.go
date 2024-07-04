package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
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
	AltEventID uuid.UUID           `json:"alternative_event_id"`
	Members    []DrvEntriesMembers `json:"members"`
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
			// log.Debug("Verein already exists: ", verein.Name)
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

		nnUuid, err := uuid.NewV7()
		if err != nil {
			return err
		}
		startberechtigt := true
		nnAthletParams := sqlc.CreateAthletParams{
			Uuid:            nnUuid,
			VereinUuid:      newVerein.Uuid,
			Name:            "Name",
			Vorname:         "No",
			Jahrgang:        "9999",
			Startberechtigt: &startberechtigt,
			Geschlecht:      "x",
		}
		_, err = crud.CreateAthlet(nnAthletParams)
		if err != nil {
			return err
		}
	}

	allNNAthleten, err := crud.GetAllNNAthleten()

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
			// log.Debugf("Rennen already exists: %s - %s", rennen.Nummer, *rennen.BezeichnungLang)
			continue
		}

		wettkampf, tag, rennabstand, err := getRennInfo(drvMeldung.Regatta.Days, r)
		if err != nil {
			return err
		}
		kosten := int32(r.Cost.Amount)

		var sex string
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
			Tag:              *tag,
			Wettkampf:        *wettkampf,
			KostenEur:        &kosten,
			Rennabstand:      rennabstand,
		}

		_, err = crud.CreateRennen(newRennen)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	allRennen, err := crud.GetAllRennen(&crud.GetAllRennenParams{
		GetMeldungen:  false,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
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
			// log.Debugf("Athlet already exists: %s %s", athlet.Vorname, athlet.Name)
			continue
		}

		startberechtigt := true

		newAthlet := sqlc.CreateAthletParams{
			Uuid:            a.Id,
			VereinUuid:      a.ClubId,
			Name:            a.Person.Firstname,
			Vorname:         a.Person.Lastname,
			Jahrgang:        a.Person.YearOfBirth,
			Startberechtigt: &startberechtigt,
			Geschlecht:      sqlc.Geschlecht(strings.ToLower(a.Person.Sex)),
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
			if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
				continue
			}

			log.Debug("Meld in DB Rev: ", meldung.DrvRevisionUuid.ClockSequence())
			log.Debug("Meld in JSON Rev: ", m.RevisionId.ClockSequence())

			if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
				retErr := api.INTERNAL_SERVER_ERROR
				retErr.Msg = fmt.Sprintf("Meldung in DB is newer than in JSON! Das sollte nicht passieren! MeldungID: %s", m.Id)
				return &retErr
			}

			// TODO: Update Meldung
			retErr := api.INTERNAL_SERVER_ERROR
			retErr.Msg = fmt.Sprintf("Min. eine Meldung in JSON is newer than in DB! Dies ist noch nicht implementiert. Bitte an Admin wenden! MeldungID: %s", m.Id)
			return &retErr
		}

		// "Default Values"
		typ := "DRV Meldung"
		bemerkung := ""
		kosten, err := getKostenForMeld(allRennen, m)
		if err != nil {
			return err
		}
		abgemeldet := false
		athleten := []crud.MeldungAthlet{}

		// Account for Alt Meldung
		// TODO: Add Col to save cor MeldUUID
		if m.AltEventID != uuid.Nil {
			log.Debug("Alternativ Meldung gefunden!")
			typ += fmt.Sprintf(" - Alternative zu RennenUUID: %s", m.AltEventID.String())
			abgemeldet = true
			*kosten = int32(0)
		}

		for _, a := range m.Members {
			role := strings.ToLower(a.Role)
			position := int32(a.Position)
			aUuid := a.ClubMemberId
			if aUuid == uuid.Nil {
				log.Warn("uuid is nil", aUuid)
				for _, nnA := range allNNAthleten {
					if nnA.VereinUuid == m.ClubId {
						aUuid = nnA.Uuid
						break
					}
				}
				if aUuid == uuid.Nil {
					log.Error("uuid is still nil", aUuid)
					return &api.INTERNAL_SERVER_ERROR
				}
			}
			if role == "cox" {
				athleten = append(athleten, crud.MeldungAthlet{
					Uuid:     aUuid,
					Position: &position,
					Rolle:    sqlc.RolleStm,
				})
			} else if role == "coach" {
				athleten = append(athleten, crud.MeldungAthlet{
					Uuid:     aUuid,
					Position: &position,
					Rolle:    sqlc.RolleTrainer,
				})
				continue
			} else if role == "rower" {
				athleten = append(athleten, crud.MeldungAthlet{
					Uuid:     aUuid,
					Position: &position,
					Rolle:    sqlc.RolleRuderer,
				})
			} else {
				log.Error("Unkown Role: ", a.Role)
				continue
			}
		}

		newMeldung := crud.CreateMeldungParams{
			CreateMeldungParams: &sqlc.CreateMeldungParams{
				Uuid:            m.Id,
				VereinUuid:      m.ClubId,
				RennenUuid:      m.EventId,
				DrvRevisionUuid: m.RevisionId,
				Abgemeldet:      &abgemeldet,
				Kosten:          *kosten,
				Typ:             typ,
				Bemerkung:       &bemerkung,
			},
			Athleten: athleten,
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

func getKostenForMeld(rennen []*crud.RennenWithMeldung, m DrvEntries) (*int32, error) {
	kosten := int32(0)

	for _, r := range rennen {
		if r.Uuid == m.EventId {
			kosten = *r.KostenEur
		}
	}

	if kosten == 0 {
		return nil, errors.New("RennenUUID von Meldung nicht gefunden!")
	}

	return &kosten, nil
}

func getRennInfo(regattaDays []string, event DrvEvents) (*sqlc.Wettkampf, *sqlc.Tag, *int32, error) {
	var (
		wettkampf   sqlc.Wettkampf
		tag         sqlc.Tag
		rennNr      int64
		rennabstand int32
		err         error
	)
	rennNr, err = strconv.ParseInt(event.Number, 10, 32)
	if err != nil {
		return nil, nil, nil, err
	}

	if event.Days[0].Date == regattaDays[0] {
		tag = sqlc.TagSa

		if rennNr < 100 {
			wettkampf = sqlc.WettkampfLangstrecke
			rennabstand = 1
		} else {
			wettkampf = sqlc.WettkampfSlalom
			if strings.Contains(event.Category.Code, "9") || strings.Contains(event.Category.Code, "10") || strings.Contains(event.Category.Code, "11") {
				rennabstand = 5
			} else if strings.Contains(event.Category.Code, "12") || strings.Contains(event.Category.Code, "13") {
				rennabstand = 4
			} else {
				rennabstand = 3
			}
		}
	} else if event.Days[0].Date == regattaDays[1] {
		tag = sqlc.TagSo

		if rennNr < 310 || rennNr == 321 {
			wettkampf = sqlc.WettkampfKurzstrecke
			rennabstand = 3
		} else {
			wettkampf = sqlc.WettkampfSlalom
			rennabstand = 10
		}
	} else {
		return nil, nil, nil, errors.New("Could not find valid Date")
	}

	return &wettkampf, &tag, &rennabstand, nil
}

// TODO: Move Func
// shuffle shuffles the elements of an array in place
func shuffle(array []*sqlc.Meldung) {
	for i := range array { //run the loop till the range of array
		j := rand.IntN(i + 1)                   //choose any random number
		array[i], array[j] = array[j], array[i] //swap the random element with current element
	}
}

func SetzungsLosung(c *fiber.Ctx) error {
	check, err := crud.CheckMeldungSetzung()
	if err != nil {
		return err
	}

	if check {
		retErr := &api.BAD_REQUEST
		retErr.Msg = "Setzung bereits erledigt! Vorher reseten um zu wiederholen!"
		return retErr
	}

	rLs, err := crud.GetAllRennen(&crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	for _, r := range rLs {
		numMeld := 0
		for _, m := range r.Meldungen {
			if *m.Abgemeldet == false {
				numMeld++
			}
		}

		abteilung := int32(1)
		bahn := int32(1)
		maxBahnen := 1

		if r.Wettkampf == sqlc.WettkampfKurzstrecke {
			maxBahnen = 4
		} else if r.Wettkampf == sqlc.WettkampfSlalom {
			maxBahnen = 3
		} else if r.Wettkampf == sqlc.WettkampfLangstrecke {
			maxBahnen = 99999
		} else if r.Wettkampf == sqlc.WettkampfStaffel {
			maxBahnen = 2
		}

		// TODO: WIP: Algo wont work correctly
		letzteVolleAbteilung := numMeld / maxBahnen
		if numMeld%maxBahnen == 1 && (r.Wettkampf != sqlc.WettkampfLangstrecke || r.Wettkampf != sqlc.WettkampfStaffel) {
			letzteVolleAbteilung--
		}

		shuffle(r.Meldungen)

		for _, m := range r.Meldungen {
			if *m.Abgemeldet {
				continue
			}

			updateParams := sqlc.UpdateMeldungSetzungParams{
				Uuid:      m.Uuid,
				Abteilung: &abteilung,
				Bahn:      &bahn,
			}
			err := crud.UpdateMeldungSetzung(updateParams)
			if err != nil {
				return err
			}

			bahn++
			if bahn > int32(maxBahnen) {
				abteilung++
				bahn = 1
			}
		}
	}

	_, err = crud.GetAllRennen(&crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	// return c.JSON(retLs)
	return c.JSON("Setzung erfolgreich erstellt!")
}

func ResetSetzung(c *fiber.Ctx) error {
	mLs, err := crud.GetAllMeldungen()
	if err != nil {
		return err
	}

	zero := int32(0)

	for _, m := range mLs {
		updateParams := sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: &zero,
			Bahn:      &zero,
		}
		err := crud.UpdateMeldungSetzung(updateParams)
		if err != nil {
			return err
		}
	}

	return c.JSON("Setzung erfolgreich zurückgesetzt!")
}

func SetStartnummern(c *fiber.Ctx) error {
	rLs, err := crud.GetAllRennen(&crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	abgStartNummer := int32(999)
	curStartNummer := int32(1)
	lastDay := sqlc.TagSa

	for _, r := range rLs {
		if lastDay != r.Tag {
			lastDay = r.Tag
			curStartNummer = 1
		}
		for _, m := range r.Meldungen {
			if *m.Abgemeldet {
				err = crud.UpdateStartNummer(sqlc.UpdateStartNummerParams{
					Uuid:        m.Uuid,
					StartNummer: &abgStartNummer,
				})
			} else {
				err = crud.UpdateStartNummer(sqlc.UpdateStartNummerParams{
					Uuid:        m.Uuid,
					StartNummer: &curStartNummer,
				})
				curStartNummer++
			}
			if err != nil {
				return err
			}
		}
	}

	return c.JSON("Startnummern erfolgreich vergeben!")
}

type SetZeitplanParams struct {
	SaStartStunde int `json:"sa_start_stunde"`
	SoStartStunde int `json:"so_start_stunde"`
}

// BUG: It is not setting time for the last race for some Reason
// TODO: Add Pausen
func SetZeitplan(c *fiber.Ctx) error {
	param := new(SetZeitplanParams)
	err := c.BodyParser(&param)
	if err != nil {
		return err
	}

	rLs, err := crud.GetAllRennen(&crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	curStartTimeSa, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SaStartStunde))
	curStartTimeSo, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SoStartStunde))

	log.Debug(curStartTimeSa, curStartTimeSo)
	log.Debug(len(rLs))
	for i, r := range rLs {
		log.Debug(i, r.Nummer, *r.Bezeichnung)
		if r.Tag == sqlc.TagSa {
			saTimeStr := curStartTimeSa.Format("15:04")

			err := crud.UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Startzeit: &saTimeStr,
				Uuid:      r.Uuid,
			})
			if err != nil {
				return err
			}

			rennenDur := time.Duration(r.NumAbteilungen*int(*r.Rennabstand)) * time.Minute
			curStartTimeSa = curStartTimeSa.Add(rennenDur)
		} else if r.Tag == sqlc.TagSo {
			soTimeStr := curStartTimeSo.Format("15:04")

			err := crud.UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Startzeit: &soTimeStr,
				Uuid:      r.Uuid,
			})
			if err != nil {
				return err
			}

			rennenDur := time.Duration(r.NumAbteilungen*int(*r.Rennabstand)) * time.Minute
			curStartTimeSo = curStartTimeSo.Add(rennenDur)
		} else {
			log.Errorf("RennenNummer %s Tag Error %s", r.Nummer, r.Tag)
		}
	}

	return c.JSON("Zeitplan erfolgreich erstellt!")
}
