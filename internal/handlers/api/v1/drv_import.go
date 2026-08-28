package api_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bata94/RegattaApi/internal/config"

	"github.com/bata94/RegattaApi/internal/crud"
	DB "github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const writeFilePerms os.FileMode = 0o666

func DrvMeldungUpload(w http.ResponseWriter, r *http.Request) {
	filename, content, err := webfw.FormFile(r, "file")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	slog.Info("Uploaded", "filename", filename)
	uploadsDir := config.C.Paths.UploadDir

	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		webfw.APIError(w, webfw.InternalError("Error while creating uploads directory"))
		return
	}

	dest := fmt.Sprintf("%s%s_%s.json", uploadsDir, "DrvMeldung", time.Now().Format("2006-01-02_15-04-05"))
	err = webfw.SaveFile(dest, content)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	err = ImportDrvJson(r.Context(), dest)
	if err != nil {
		webfw.APIError(w, webfw.InternalError("An Error occurred while importing the JSON File! If you directly downloaded the File from DRV and uploaded it, without modifying it, please contact the Admin! Details: "+err.Error()))
		return
	}

	if webfw.IsHtmxRequest(r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "File uploaded successfully!"}); err != nil {
		slog.Warn("JSON encode error", "err", err)
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
	slog.Info("Importing Drv JSON File", "file", filePath)

	b, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("Read File Error", "err", err)
		return err
	}

	drvMeldung := DrvMeldungJson{}
	err = json.Unmarshal(b, &drvMeldung)
	if err != nil {
		slog.Error("Unmarshal Error", "err", err)
		return err
	}

	o, err := json.MarshalIndent(drvMeldung, "", "  ")
	if err != nil {
		slog.Error("Marshal Error", "err", err)
		return err
	}
	if err := os.WriteFile(fmt.Sprintf("%sImported_DrvMeldung_%s.json", config.C.Paths.UploadDir, time.Now().Format("15-04-05")), o, writeFilePerms); err != nil {
		slog.Error("Error writing debug JSON", "err", err)
	}

	return DB.WithTx(ctx, func(txCtx context.Context) error {
		return importDrvJsonCore(txCtx, drvMeldung)
	})
}

func importDrvJsonCore(ctx context.Context, drvMeldung DrvMeldungJson) error {
	for _, v := range drvMeldung.Clubs {
		verein, err := crud.GetVereinMinimal(ctx, v.Id)
		if err != nil {
			if err != apierr.ErrNotFound {
				slog.Error("Crud get Error", "err", err)
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
			slog.Error("Crud create Error", "err", err)
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

	allNNAthleten, err := crud.GetAllNNAthleten(ctx)
	if err != nil {
		return err
	}

	for _, r := range drvMeldung.Events {
		rennen, err := crud.GetRennenMinimal(ctx, r.Id)
		if err != nil {
			if err != apierr.ErrNotFound {
				slog.Error("Crud get Error", "err", err)
				return err
			}
			if rennen.Uuid != uuid.Nil {
				slog.Debug("Rennen already exists", "nummer", rennen.Nummer, "bezeichnung", rennen.BezeichnungLang)
				continue
			}
		}
		if rennen.Uuid != uuid.Nil {
			slog.Debug("Rennen already exists", "nummer", rennen.Nummer, "bezeichnung", rennen.BezeichnungLang)
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

		sortOrder := int32(r.Days[0].SortOrder)

		if r.Days[0].Date == "2024-09-29" {
			slog.Debug("SortOrder Sonntag!")
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
			Geschlecht:       sqlc.Geschlecht(sex),
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
			slog.Error("Crud create Error", "err", err)
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
	slog.Debug("Len All Rennen", "count", len(allRennen))

	for _, a := range drvMeldung.ClubMembers {
		athlet, err := crud.GetAthletMinimal(ctx, a.Id)
		if err != nil {
			if err != apierr.ErrNotFound {
				slog.Error("Crud get Error", "err", err)
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
			Geschlecht:      sqlc.Geschlecht(strings.ToLower(a.Person.Sex)),
		}

		_, err = crud.CreateAthlet(ctx, newAthlet)
		if err != nil {
			slog.Error("Crud create Error", "err", err)
			return err
		}
	}

	slog.Debug("Import Entries Loop...")
	for _, m := range drvMeldung.Entries {
		meldung, err := crud.GetMeldungMinimal(ctx, m.Id)
		if err != nil {
			if err != apierr.ErrNotFound {
				slog.Error("Crud get Error", "err", err)
				return err
			}
			if meldung.Uuid != uuid.Nil {
				if meldung.DrvRevisionUuid.Time() == m.RevisionId.Time() {
					slog.Debug("Meldung exists in DB, skipping...")
					continue
				}

				slog.Debug("MeldUuid", "uuid", meldung.Uuid)
				slog.Debug("Meld in DB Rev", "rev", meldung.DrvRevisionUuid.Time())
				slog.Debug("Meld in JSON Rev", "rev", m.RevisionId.Time())

				if meldung.DrvRevisionUuid.Time() > m.RevisionId.Time() {
					slog.Debug("Meldung in DB is newer than in JSON", "meldungID", m.Id)
					continue
				}

				slog.Debug("Min. eine Meldung in JSON is newer than in DB", "meldungID", m.Id)
				continue
			}
		}
		if meldung.Uuid != uuid.Nil {
			if meldung.DrvRevisionUuid.Time() == m.RevisionId.Time() {
				slog.Debug("Meldung exists in DB, skipping...")
				continue
			}

			slog.Debug("MeldUuid", "uuid", meldung.Uuid)
			slog.Debug("Meld in DB Rev", "rev", meldung.DrvRevisionUuid.Time())
			slog.Debug("Meld in JSON Rev", "rev", m.RevisionId.Time())

			if meldung.DrvRevisionUuid.Time() > m.RevisionId.Time() {
				slog.Debug("Meldung in DB is newer than in JSON", "meldungID", m.Id)
				continue
			}

			slog.Debug("Min. eine Meldung in JSON is newer than in DB", "meldungID", m.Id)
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
			slog.Debug("Alternativ Meldung gefunden", "status", m.Status)
			typ += fmt.Sprintf(" - Alternative zu RennenUUID: %s", m.AltEventID.String())
			abgemeldet = true
			kosten = int32(0)
		}

		for _, a := range m.Members {
			role := strings.ToLower(a.Role)
			aUuid := a.ClubMemberId
			if aUuid == uuid.Nil {
				slog.Debug("uuid is nil", "uuid", aUuid)
				for _, nnA := range allNNAthleten {
					if nnA.VereinUuid == m.ClubId {
						aUuid = nnA.Uuid
						break
					}
				}
				if aUuid == uuid.Nil {
					slog.Warn("uuid is still nil", "uuid", aUuid)
					return webfw.InternalError("Failed to find athlete UUID")
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
				slog.Warn("Unknown Role", "role", a.Role)
				continue
			}
			athleten = append(athleten, crud.CreateMeldungAthletParams{
				Uuid:     aUuid,
				Position: int32(a.Position),
				Rolle:    rolle,
			})
		}

		slog.Debug("Members done... Creating Meldung")
		newMeldung := crud.CreateMeldungParams{
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
			Athleten:        athleten,
		}

		_, err = crud.CreateMeldung(ctx, newMeldung)
		if err != nil {
			slog.Error("Crud create Error", "err", err)
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
		slog.Error("rennen UUID von meldung nicht gefunden", "eventId", m.EventId)
		return 0, errors.New("rennenUUID von meldung nicht gefunden")
	}

	return kosten, nil
}

const (
	slalomRennNrThreshold     = 100
	staffelRennNrThreshold    = 310
	specialStaffelRennNr      = 321
	langstreckeAbstand        = 1
	slalomAlter9bis11Abstand  = 5
	slalomAlter12bis13Abstand = 4
	slalomAb14Abstand         = 3
	kurzstreckeAbstand        = 3
	staffelAbstand            = 10
)

func getRennInfo(regattaDays []string, event DrvEvents) (*sqlc.Wettkampf, *sqlc.Tag, int32, error) {
	var (
		wettkampf   sqlc.Wettkampf
		tag         sqlc.Tag
		rennNr      int64
		rennabstand int32
		err         error
	)
	rennNr, err = strconv.ParseInt(event.Number, 10, 32)
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
			if strings.Contains(event.Category.Code, "9") || strings.Contains(event.Category.Code, "10") || strings.Contains(event.Category.Code, "11") {
				rennabstand = slalomAlter9bis11Abstand
			} else if strings.Contains(event.Category.Code, "12") || strings.Contains(event.Category.Code, "13") {
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
		return nil, nil, 0, errors.New("could not find valid date")
	}

	return &wettkampf, &tag, rennabstand, nil
}
