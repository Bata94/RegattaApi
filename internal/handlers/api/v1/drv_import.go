package api_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bata94/RegattaApi/internal/config"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const writeFilePerms os.FileMode = 0o666

func DrvMeldungUpload(c *handler.Context) error {
	filename, content, err := c.FormFile("file")
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	fmt.Println("Uploaded:", filename)
	uploadsDir := config.C.Paths.UploadDir

	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		return handler.InternalError("Error while creating uploads directory")
	}

	dest := fmt.Sprintf("%s%s_%s.json", uploadsDir, "DrvMeldung", time.Now().Format("2006-01-02_15-04-05"))
	err = c.SaveFile(dest, content)
	if err != nil {
		return handler.InternalError(err.Error())
	}

	err = ImportDrvJson(c.Request.Context(), dest)
	if err != nil {
		return handler.InternalError("An Error occurred while importing the JSON File! If you directly downloaded the File from DRV and uploaded it, without modifying it, please contact the Admin! Details: " + err.Error())
	}

	if c.IsHtmxRequest() {
		return nil
	} else {
		return c.JSON("File uploaded successfully!")
	}
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
	Id         uuid.UUID           `json:"id"`
	RevisionId uuid.UUID           `json:"revision_id"`
	EventId    uuid.UUID           `json:"event_id"`
	ClubId     uuid.UUID           `json:"club_id"`
	Name       string              `json:"name"`
	ShortName  string              `json:"shortname"`
	Sequence   int                 `json:"sequence"`
	Status     int                 `json:"status"`
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

func ImportDrvJson(ctx context.Context, filePath string) error {
	fmt.Println("Importing Drv JSON File:", filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Read File Error:", err)
		return err
	}

	drvMeldung := DrvMeldungJson{}
	err = json.Unmarshal(b, &drvMeldung)
	if err != nil {
		fmt.Println("Unmarshal Error:", err)
		return err
	}

	o, err := json.MarshalIndent(drvMeldung, "", "  ")
	if err != nil {
		fmt.Println("Marshal Error:", err)
		return err
	}
	os.WriteFile(fmt.Sprintf("%sImported_DrvMeldung_%s.json", config.C.Paths.UploadDir, time.Now().Format("15-04-05")), o, writeFilePerms)

	// TODO: Use a Transaction here!

	for _, v := range drvMeldung.Clubs {
		verein, err := crud.GetVereinMinimal(ctx, v.Id)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				fmt.Println( "Crud get Error:", err)
				return err
			}
			if verein.Uuid != uuid.Nil {
				continue
			}
		}
		if verein.Uuid != uuid.Nil {
			continue
		}

		newVerein := sqlc.CreateVereinParams{
			Uuid:     v.Id,
			Name:     v.Name,
			Kurzform: v.ShortName,
			Kuerzel:  v.Code,
		}
		_, err = crud.CreateVerein(ctx, newVerein)
		if err != nil {
			fmt.Println( "Crud create Error:", err)
			return err
		}

		nnUuid, err := uuid.NewV7()
		if err != nil {
			return err
		}
		nnAthletParams := sqlc.CreateAthletParams{
			Uuid:            nnUuid,
			VereinUuid:      newVerein.Uuid,
			Name:            "Name",
			Vorname:         "No",
			Jahrgang:        "9999",
			Startberechtigt: false,
			Geschlecht:      "x",
		}
		_, err = crud.CreateAthlet(ctx, nnAthletParams)
		if err != nil {
			return err
		}
	}

	allNNAthleten, err := crud.GetAllNNAthleten(ctx, )
	if err != nil {
		return err
	}

	for _, r := range drvMeldung.Events {
		rennen, err := crud.GetRennenMinimal(ctx, r.Id)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				fmt.Println( "Crud get Error:", err)
				return err
			}
			if rennen.Uuid != uuid.Nil {
				fmt.Printf( "Rennen already exists: %s - %s\n", rennen.Nummer, rennen.BezeichnungLang)
				continue
			}
		}
		if rennen.Uuid != uuid.Nil {
			fmt.Printf( "Rennen already exists: %s - %s\n", rennen.Nummer, rennen.BezeichnungLang)
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
			sex = strings.ToLower( r.Sex)
		}

		sortOrder := int32(r.Days[0].SortOrder)

		if r.Days[0].Date == "2024-09-29" {
			fmt.Println( "SortOrder Sonntag!")
			sortOrder += 500
		}

		newRennen := sqlc.CreateRennenParams{
			Uuid:             r.Id,
			SortID:           sortOrder,
			Nummer:           r.Number,
			Bezeichnung:      r.Code,
			BezeichnungLang:  r.Name,
			Zusatz:           pgtype.Text{String: r.Addition, Valid: true},
			Leichtgewicht:    r.Weighed,
			Geschlecht:       sqlc.Geschlecht( sex),
			Bootsklasse:      r.BoatType.Code,
			BootsklasseLang:  r.BoatType.Name,
			Altersklasse:     r.Category.Code,
			AltersklasseLang: r.Category.Name,
			Tag:              *tag,
			Wettkampf:        *wettkampf,
			KostenEur:        pgtype.Int4{Int32: kosten, Valid: true},
			Rennabstand:      pgtype.Int4{Int32: rennabstand, Valid: true},
		}

		_, err = crud.CreateRennen(ctx, newRennen)
		if err != nil {
			fmt.Println( "Crud create Error:", err)
			return err
		}
	}

	allRennen, err := crud.GetAllRennen(ctx, crud.GetAllRennenParams{
		GetMeldungen:  false,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}
	fmt.Println( "Len All Rennen:", len(allRennen))

	for _, a := range drvMeldung.ClubMembers {
		athlet, err := crud.GetAthletMinimal(ctx, a.Id)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				fmt.Println( "Crud get Error:", err)
				return err
			}
			if athlet.Uuid != uuid.Nil {
				continue
			}
		}
		if athlet.Uuid != uuid.Nil {
			continue
		}

		startberechtigt := true
		newAthlet := sqlc.CreateAthletParams{
			Uuid:            a.Id,
			VereinUuid:      a.ClubId,
			Name:            a.Person.Lastname,
			Vorname:         a.Person.Firstname,
			Jahrgang:        a.Person.YearOfBirth,
			Startberechtigt: startberechtigt,
			Geschlecht:      sqlc.Geschlecht( strings.ToLower(a.Person.Sex)),
		}

		_, err = crud.CreateAthlet(ctx, newAthlet)
		if err != nil {
			fmt.Println( "Crud create Error:", err)
			return err
		}
	}

	fmt.Println( "Import Entries Loop...")
	for _, m := range drvMeldung.Entries {
		meldung, err := crud.GetMeldungMinimal(ctx, m.Id)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				fmt.Println( "Crud get Error:", err)
				return err
			}
			if meldung.Uuid != uuid.Nil {
				if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
					fmt.Println( "Meldung exists in DB, skipping...")
					continue
				}

				fmt.Println( "MeldUuid:", meldung.Uuid)
				fmt.Println( "Meld in DB Rev:", meldung.DrvRevisionUuid.ClockSequence())
				fmt.Println( "Meld in JSON Rev:", m.RevisionId.ClockSequence())

				if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
					fmt.Printf( "Meldung in DB is newer than in JSON! MeldungID: %s\n", m.Id)
					continue
				}

				fmt.Printf( "Min. eine Meldung in JSON is newer than in DB! MeldungID: %s\n", m.Id)
				continue
			}
		}
		if meldung.Uuid != uuid.Nil {
			if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
				fmt.Println( "Meldung exists in DB, skipping...")
				continue
			}

			fmt.Println( "MeldUuid:", meldung.Uuid)
			fmt.Println( "Meld in DB Rev:", meldung.DrvRevisionUuid.ClockSequence())
			fmt.Println( "Meld in JSON Rev:", m.RevisionId.ClockSequence())

			if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
				fmt.Printf( "Meldung in DB is newer than in JSON! MeldungID: %s\n", m.Id)
				continue
			}

			fmt.Printf( "Min. eine Meldung in JSON is newer than in DB! MeldungID: %s\n", m.Id)
			continue
		}

		typ := "DRV Meldung"
		bemerkung := ""
		kosten, err := getKostenForMeld(allRennen, m)
		if err != nil {
			return err
		}
		abgemeldet := false
		athleten := []crud.CreateMeldungAthletParams{}

		if m.AltEventID != uuid.Nil {
			fmt.Println( "Alternativ Meldung gefunden! Status:", m.Status)
			typ += fmt.Sprintf( " - Alternative zu RennenUUID: %s", m.AltEventID.String())
			abgemeldet = true
			kosten = int32(0)
		}

		for _, a := range m.Members {
			role := strings.ToLower( a.Role)
			aUuid := a.ClubMemberId
			if aUuid == uuid.Nil {
				fmt.Println( "uuid is nil", aUuid)
				for _, nnA := range allNNAthleten {
					if nnA.VereinUuid == m.ClubId {
						aUuid = nnA.Uuid
						break
					}
				}
				if aUuid == uuid.Nil {
					fmt.Println( "uuid is still nil", aUuid)
					return handler.InternalError("Failed to find athlete UUID")
				}
			}
			var rolle sqlc.Rolle
			if role == "cox" {
				rolle = sqlc.RolleStm
			} else if role == "coach" {
				rolle = sqlc.RolleTrainer
				continue
			} else if role == "rower" {
				rolle = sqlc.RolleRuderer
			} else {
				fmt.Println( "Unknown Role:", a.Role)
				continue
			}
			athleten = append(athleten, crud.CreateMeldungAthletParams{
				Uuid:     aUuid,
				Position: int32(a.Position),
				Rolle:    rolle,
			})
		}

		fmt.Println( "Members done... Creating Meldung")
		newMeldung := crud.CreateMeldungParams{
			CreateMeldungParams: sqlc.CreateMeldungParams{
				Uuid:            m.Id,
				VereinUuid:      m.ClubId,
				RennenUuid:      m.EventId,
				DrvRevisionUuid: m.RevisionId,
				Abgemeldet:      abgemeldet,
				StartNummer:     int32(0),
				Abteilung:       int32(0),
				Bahn:            int32(0),
				Kosten:          kosten,
				Typ:             typ,
				Bemerkung:       pgtype.Text{String: bemerkung},
			},
			Athleten: athleten,
		}

		_, err = crud.CreateMeldung(ctx, newMeldung)
		if err != nil {
			fmt.Println( "Crud create Error", err)
			return err
		}
	}

	return nil
}

func getKostenForMeld(rennen []crud.Rennen, m DrvEntries) (int32, error) {
	kosten := int32(0)

	for _, r := range rennen {
		if r.Uuid == m.EventId {
			if v := r.GetKostenEur(); v != nil {
				kosten = int32(*v)
			}
		}
	}

	if kosten == 0 {
		fmt.Println( m.EventId)
		return 0, errors.New( "RennenUUID von Meldung nicht gefunden!")
	}

	return kosten, nil
}

const (
	slalomRennNrThreshold    = 100
	staffelRennNrThreshold   = 310
	specialStaffelRennNr     = 321
	langstreckeAbstand       = 1
	slalomAlter9bis11Abstand = 5
	slalomAlter12bis13Abstand = 4
	slalomAb14Abstand        = 3
	kurzstreckeAbstand       = 3
	staffelAbstand           = 10
)

func getRennInfo(regattaDays []string, event DrvEvents) (*sqlc.Wettkampf, *sqlc.Tag, int32, error) {
	var (
		wettkampf   sqlc.Wettkampf
		tag         sqlc.Tag
		rennNr      int64
		rennabstand int32
		err         error
	)
	rennNr, err = strconv.ParseInt( event.Number, 10, 32)
	if err != nil {
		return nil, nil, 0, err
	}

	switch event.Days[0].Date {
	case regattaDays[0]:
		tag = sqlc.TagSa

		if rennNr < slalomRennNrThreshold {
			wettkampf = sqlc.WettkampfLangstrecke
			rennabstand = langstreckeAbstand
		} else {
			wettkampf = sqlc.WettkampfSlalom
			if strings.Contains( event.Category.Code, "9") || strings.Contains( event.Category.Code, "10") || strings.Contains( event.Category.Code, "11") {
				rennabstand = slalomAlter9bis11Abstand
			} else if strings.Contains( event.Category.Code, "12") || strings.Contains( event.Category.Code, "13") {
				rennabstand = slalomAlter12bis13Abstand
			} else {
				rennabstand = slalomAb14Abstand
			}
		}
	case regattaDays[1]:
		tag = sqlc.TagSo

		if rennNr < staffelRennNrThreshold || rennNr == specialStaffelRennNr {
			wettkampf = sqlc.WettkampfKurzstrecke
			rennabstand = kurzstreckeAbstand
		} else {
			wettkampf = sqlc.WettkampfStaffel
			rennabstand = staffelAbstand
		}
	default:
		return nil, nil, 0, errors.New( "Could not find valid Date")
	}

	return &wettkampf, &tag, rennabstand, nil
}
