package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/internal/templates/pdf"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

func GetPdfFooter(c *fiber.Ctx) error {
	return handlers.RenderPdf(c, "footer", pdf_templates.PdfFooter())
}

func GetMeldeergebnisList(c *fiber.Ctx) error {
	files, err := utils.GetFilenames("meldeergebnis")
	if err != nil {
		return err
	}
	return api.JSON(c, files)
}

func GetMeldeergebnisFilename(c *fiber.Ctx) error {
	filename := c.Params("filename")
	return c.SendFile(filepath.Join("./files", "meldeergebnis", filename), true)
}

func GetMeldeergebnisHtml(c *fiber.Ctx) error {
	pLs, err := crud.GetAllPausen()
	if err != nil {
		return err
	}
	rLs, err := crud.GetAllRennenWithAthlet(crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	pLsParsed := []pdf_templates.PausenMeldeergebnisPDF{}
	for _, p := range pLs {
		pLsParsed = append(pLsParsed, pdf_templates.PausenMeldeergebnisPDF{
			Id:             int(p.ID),
			Laenge:         int(p.Laenge),
			NachRennenUuid: p.NachRennenUuid.String(),
		})
	}

	rLsParsed := []pdf_templates.RennenMeldeergebnisPDF{}
	for _, r := range rLs {
		rParsed := pdf_templates.RennenMeldeergebnisPDF{
			Uuid:              r.Uuid.String(),
			RennNr:            r.Nummer,
			Bezeichnung:       r.Bezeichnung,
			BezeichnungZusatz: r.Zusatz,
			Startzeit:         r.Startzeit,
			Rennabstand:       r.Rennabstand,
			Tag:               string(r.Tag),
			NumMeldungen:      *r.NumMeldungen,
			NumAbteilungen:    *r.NumAbteilungen,
			Wettkampf:         r.Wettkampf,
			Abteilungen:       make([]pdf_templates.AbteilungenMeldeergebnisPDF, *r.NumAbteilungen),
			Abmeldungen:       []pdf_templates.MeldungMeldeergebnisPDF{},
		}

		for i := range rParsed.Abteilungen {
			rParsed.Abteilungen[i].Nummer = i + 1
		}

		if len(r.Meldungen) == 0 {
			rLsParsed = append(rLsParsed, rParsed)
			continue
		}
		for _, m := range r.Meldungen {
			athletenStr := ""
			for _, a := range m.Athleten {
				if *a.Rolle == sqlc.RolleTrainer {
					continue
				}

				if athletenStr != "" {
					athletenStr += ", "
				}

				if *a.Rolle == sqlc.RolleStm {
					athletenStr += fmt.Sprintf("\nStm.: %s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
				} else {
					athletenStr += fmt.Sprintf("%s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
				}
			}

			meldungEntry := pdf_templates.MeldungMeldeergebnisPDF{
				StartNummer: int(m.StartNummer),
				Bahn:        int(m.Bahn),
				Teilnehmer:  athletenStr,
				Verein:      m.Verein.Name,
			}

			if m.Abgemeldet {
				rParsed.Abmeldungen = append(rParsed.Abmeldungen, meldungEntry)
				continue
			}

			abteilung := int(m.Abteilung)
			mParsed := meldungEntry
			log.Debugf("RennNr %s numAbt %d lenAbt %d curAbt %d", r.Nummer, r.NumAbteilungen, len(rParsed.Abteilungen), abteilung)
			// BUG: Throws Error if Setzung not done
			rParsed.Abteilungen[abteilung-1].Meldungen = append(rParsed.Abteilungen[abteilung-1].Meldungen, mParsed)
		}

		rLsParsed = append(rLsParsed, rParsed)
	}

	return handlers.RenderPdf(
		c,
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		pdf_templates.MeldeErgebnis(rLsParsed, pLsParsed),
	)
}

func GenerateMeldeergebnis(c *fiber.Ctx) error {
	filepath, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	return c.SendFile(filepath, true)
}

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
		return &api.ReqError{
			Code:       500,
			StatusCode: fiber.StatusInternalServerError,
			Title:      api.INTERNAL_SERVER_ERROR.Title,
			Msg:        "An Error accurred while importing the JSON File! If you directly downloaded the File from DRV and uploaded it, without modifying it, please contact the Admin!",
			Details:    err.Error(),
			Data:       nil,
		}
	}

	return api.JSON(c, "File uploaded successfully!")
}

func GenerateErgebnisHtml(c *fiber.Ctx) error {
	rLsParsed := []pdf_templates.ErgebnisRennenPDF{}
	rennen, err := crud.GetAllRennenWithAthlet(crud.GetAllRennenParams{
		GetMeldungen:  true,
		GetAthleten:   true,
		ShowEmpty:     false,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	for _, r := range rennen {
    if r.Wettkampf != sqlc.WettkampfLangstrecke {
      break
    }
		if *r.NumMeldungen != 0 {
      rParsed := pdf_templates.ErgebnisRennenPDF{
        Uuid:              r.Uuid.String(),
        RennNr:            r.Nummer,
        Bezeichnung:       r.Bezeichnung,
        BezeichnungZusatz: r.Zusatz,
        Startzeit:         r.Startzeit,
        Rennabstand:       r.Rennabstand,
        Tag:               string(r.Tag),
        NumMeldungen:      *r.NumMeldungen,
        NumAbteilungen:    *r.NumAbteilungen,
        Wettkampf:         r.Wettkampf,
        Abteilungen:       make([]pdf_templates.ErgebnisAbteilungPDF, *r.NumAbteilungen),
        Dns:       []pdf_templates.MeldungMeldeergebnisPDF{},
      }

      for i := range rParsed.Abteilungen {
        rParsed.Abteilungen[i].Nummer = i + 1
      }

      for _, m := range r.Meldungen {
        if m.Abgemeldet {
          continue
        }

        athletenStr := ""
        for _, a := range m.Athleten {
          if *a.Rolle == sqlc.RolleTrainer {
            continue
          }

          if athletenStr != "" {
            athletenStr += ", "
          }

          if *a.Rolle == sqlc.RolleStm {
            athletenStr += fmt.Sprintf("\nStm.: %s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
          } else {
            athletenStr += fmt.Sprintf("%s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
          }
        }

        ergebnis, err := crud.GetZeitnahmeErgebnisByMeld(m.Uuid)
        if err != nil {
          rParsed.Dns = append(rParsed.Dns, pdf_templates.MeldungMeldeergebnisPDF{
            StartNummer: int(m.StartNummer),
            Bahn:        int(m.Bahn),
            Teilnehmer:  athletenStr,
            Verein:      m.Verein.Name,
          })
          continue
        }

        endZeit := time.Duration(ergebnis.Endzeit * float64(time.Second))
        minutes := int(endZeit / time.Minute)
        secondsPart := int((endZeit % time.Minute) / time.Second)
        milliseconds := int((endZeit % time.Second) / time.Millisecond)

        // Print the formatted time as "Minutes:Seconds.Milliseconds"
        endZeitStr := fmt.Sprintf("%02d:%02d.%03d\n", minutes, secondsPart, milliseconds)

        meldungEntry := pdf_templates.ErgebnisMeldungPDF{
          StartNummer: int(m.StartNummer),
          Bahn:        int(m.Bahn),
          Teilnehmer:  athletenStr,
          Verein:      m.Verein.Name,
          Platz:      1,
          Endzeit:     ergebnis.Endzeit,
          EndzeitStr: endZeitStr,
        }

        abteilung := int(m.Abteilung)
        mParsed := meldungEntry
        log.Debugf("RennNr %s numAbt %d lenAbt %d curAbt %d", r.Nummer, r.NumAbteilungen, len(rParsed.Abteilungen), abteilung)
        // BUG: Throws Error if Setzung not done
        rParsed.Abteilungen[abteilung-1].Meldungen = append(rParsed.Abteilungen[abteilung-1].Meldungen, mParsed)

      }

      for i, abt := range rParsed.Abteilungen {
        sort.Slice(abt.Meldungen, func(i, j int) bool {
          return abt.Meldungen[i].Endzeit < abt.Meldungen[j].Endzeit
        })

        p := 1

        for j, _ := range abt.Meldungen {
          rParsed.Abteilungen[i].Meldungen[j].Platz = p
          p ++
        }
      }
      rLsParsed = append(rLsParsed, rParsed)
		}
	}

	return handlers.RenderPdf(
		c,
		fmt.Sprintf("Ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		pdf_templates.Ergebnis(rLsParsed),
	)
}

func GenerateErgebnis(c *fiber.Ctx) error {
	filepath, err := utils.SavePDFfromHTML(
		"leitung/ergebnis",
		"ergebnis",
		fmt.Sprintf("ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	return c.SendFile(filepath, true)
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
	Status     int       `json:"status"` // 1 default Meldung, 8 Abmeldung, 256 Alternativ Meldung
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
		if verein.Uuid != uuid.Nil {
			continue
		}

		newVerein := sqlc.CreateVereinParams{
			Uuid:     v.Id,
			Name:     v.Name,
			Kurzform: v.ShortName,
			Kuerzel:  v.Code,
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
		nnAthletParams := sqlc.CreateAthletParams{
			Uuid:            nnUuid,
			VereinUuid:      newVerein.Uuid,
			Name:            "Name",
			Vorname:         "No",
			Jahrgang:        "9999",
			Startberechtigt: false,
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
		if rennen.Uuid != uuid.Nil {
			log.Debugf("Rennen already exists: %s - %s", rennen.Nummer, rennen.BezeichnungLang)
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

		// TODO: Quickfix pre import!
		sortOrder := int32(r.Days[0].SortOrder)

		if r.Days[0].Date == "2024-09-29" {
			log.Debug("SortOrder Sonntag!")
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

		_, err = crud.CreateRennen(newRennen)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	allRennen, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  false,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}
	log.Debug("Len All Rennen: ", len(allRennen))

	for _, a := range drvMeldung.ClubMembers {
		athlet, err := crud.GetAthletMinimal(a.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
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

		_, err = crud.CreateAthlet(newAthlet)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	log.Debug("Import Entries Loop...")
	for _, m := range drvMeldung.Entries {
		meldung, err := crud.GetMeldungMinimal(m.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				log.Error("Crud get Error")
				log.Error(err.Error())
				return err
			}
		}
		if meldung.Uuid != uuid.Nil {
			if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
				log.Debug("Meldung exists in DB, skipping...")
				continue
			}

			log.Debug("MeldUuid: ", meldung.Uuid)
			log.Debug("Meld in DB Rev: ", meldung.DrvRevisionUuid.ClockSequence())
			log.Debug("Meld in JSON Rev: ", m.RevisionId.ClockSequence())

			if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
				retErr := api.INTERNAL_SERVER_ERROR
				retErr.Msg = fmt.Sprintf("Meldung in DB is newer than in JSON! Das sollte nicht passieren! MeldungID: %s", m.Id)
				log.Error(retErr)
				continue
			}

			// TODO: Update Meldung
			retErr := api.INTERNAL_SERVER_ERROR
			retErr.Msg = fmt.Sprintf("Min. eine Meldung in JSON is newer than in DB! Dies ist noch nicht implementiert. Bitte an Admin wenden! MeldungID: %s", m.Id)
			log.Error(retErr)
			continue
		}

		// "Default Values"
		typ := "DRV Meldung"
		bemerkung := ""
		kosten, err := getKostenForMeld(allRennen, m)
		if err != nil {
			return err
		}
		abgemeldet := false
		athleten := []crud.CreateMeldungAthletParams{}

		// Account for Alt Meldung
		// TODO: Add Col to save cor MeldUUID
		if m.AltEventID != uuid.Nil {
			log.Debug("Alternativ Meldung gefunden! Status: ", m.Status)
			typ += fmt.Sprintf(" - Alternative zu RennenUUID: %s", m.AltEventID.String())
			abgemeldet = true
			kosten = int32(0)
		}

		for _, a := range m.Members {
			role := strings.ToLower(a.Role)
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
			var rolle sqlc.Rolle
			if role == "cox" {
				rolle = sqlc.RolleStm
			} else if role == "coach" {
				rolle = sqlc.RolleTrainer
				continue
			} else if role == "rower" {
				rolle = sqlc.RolleRuderer
			} else {
				log.Error("Unkown Role: ", a.Role)
				continue
			}
			athleten = append(athleten, crud.CreateMeldungAthletParams{
				Uuid:     aUuid,
				Position: int32(a.Position),
				Rolle:    rolle,
			})
		}

		log.Debug("Members done... Creating Meldung")
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

		_, err = crud.CreateMeldung(newMeldung)
		if err != nil {
			log.Error("Crud create Error")
			log.Error(err.Error())
			return err
		}
	}

	return nil
}

func getKostenForMeld(rennen []crud.Rennen, m DrvEntries) (int32, error) {
	kosten := int32(0)

	for _, r := range rennen {
		if r.Uuid == m.EventId {
			kosten = int32(r.KostenEur)
		}
	}

	if kosten == 0 {
		log.Error(m.EventId)
		return 0, errors.New("RennenUUID von Meldung nicht gefunden!")
	}

	return kosten, nil
}

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
			wettkampf = sqlc.WettkampfStaffel
			rennabstand = 10
		}
	} else {
		return nil, nil, 0, errors.New("Could not find valid Date")
	}

	return &wettkampf, &tag, rennabstand, nil
}

// TODO: Move Func
func shuffle(array []crud.Meldung) []crud.Meldung {
	for i := range array {
		j := rand.IntN(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
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

	rLs, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	for _, r := range rLs {
		numMeld := r.NumMeldungen
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
		letzteVolleAbteilung := *numMeld / maxBahnen
		if *numMeld%maxBahnen == 1 && (r.Wettkampf != sqlc.WettkampfLangstrecke || r.Wettkampf != sqlc.WettkampfStaffel) {
			letzteVolleAbteilung--
		}

		r.Meldungen = shuffle(r.Meldungen)

		for _, m := range r.Meldungen {
			if m.Abgemeldet {
				continue
			}
			log.Debugf("Setzung: %s Abt: %d Bahn: %d", m.Uuid, abteilung, bahn)
			if err := crud.UpdateMeldungSetzung(sqlc.UpdateMeldungSetzungParams{
				Uuid:      m.Uuid,
				Abteilung: abteilung,
				Bahn:      bahn,
			}); err != nil {
				return err
			}
			bahn++
			if bahn > int32(maxBahnen) {
				abteilung++
				bahn = 1
			}
		}
	}
	return api.JSON(c, "Setzung erfolgreich erstellt!")
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
			Abteilung: zero,
			Bahn:      zero,
		}
		err := crud.UpdateMeldungSetzung(updateParams)
		if err != nil {
			return err
		}
	}

	return api.JSON(c, "Setzung erfolgreich zurückgesetzt!")
}

func SetStartnummern(c *fiber.Ctx) error {
	rLs, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	abgStartNummer := int32(-999)
	curStartNummer := int32(1)
	lastDay := sqlc.TagSa

	for _, r := range rLs {
		if lastDay != r.Tag {
			lastDay = r.Tag
			curStartNummer = 1
		}
		for _, m := range r.Meldungen {
			if m.Abgemeldet {
				err = crud.UpdateStartNummer(sqlc.UpdateStartNummerParams{
					Uuid:        m.Uuid,
					StartNummer: abgStartNummer,
				})
			} else {
				err = crud.UpdateStartNummer(sqlc.UpdateStartNummerParams{
					Uuid:        m.Uuid,
					StartNummer: curStartNummer,
				})
				curStartNummer++
			}
			if err != nil {
				return err
			}
		}
	}

	return api.JSON(c, "Startnummern erfolgreich vergeben!")
}

type SetZeitplanParams struct {
	SaStartStunde int `json:"sa_start_stunde"`
	SoStartStunde int `json:"so_start_stunde"`
}

func SetZeitplan(c *fiber.Ctx) error {
	param := new(SetZeitplanParams)
	err := c.BodyParser(&param)
	if err != nil {
		return err
	}

	rLs, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	pLs, err := crud.GetAllPausen()
	if err != nil {
		return err
	}

	curStartTimeSa, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SaStartStunde))
	curStartTimeSo, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SoStartStunde))

	for _, r := range rLs {
		if r.Tag == sqlc.TagSa {
			saTimeStr := curStartTimeSa.Format("15:04")

			log.Debugf("Setting RennenNr: %s to time %s", r.Nummer, saTimeStr)
			err := crud.UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Startzeit: pgtype.Text{String: saTimeStr, Valid: true},
				Uuid:      r.Uuid,
			})
			if err != nil {
				return err
			}

			rennenDur := time.Duration(*r.NumAbteilungen*int(r.Rennabstand)) * time.Minute
			if r.Wettkampf == sqlc.WettkampfLangstrecke {
				rennenDur = time.Duration(*r.NumMeldungen*int(r.Rennabstand)) * time.Minute
			}
			curStartTimeSa = curStartTimeSa.Add(rennenDur)

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					pausenDur := time.Duration(p.Laenge) * time.Minute
					curStartTimeSa = curStartTimeSa.Add(pausenDur)

					curMinuteStr := fmt.Sprint(curStartTimeSa.Minute())
					curMinute, err := strconv.Atoi(curMinuteStr[1:])
					if err != nil {
						log.Error(err)
						continue
					}

					roundingMinutes := 10 - curMinute
					roundingDur := time.Duration(roundingMinutes) * time.Minute
					curStartTimeSa = curStartTimeSa.Add(roundingDur)
				}
			}
		} else if r.Tag == sqlc.TagSo {
			soTimeStr := curStartTimeSo.Format("15:04")

			err := crud.UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Startzeit: pgtype.Text{String: soTimeStr, Valid: true},
				Uuid:      r.Uuid,
			})
			if err != nil {
				return err
			}

			rennenDur := time.Duration(*r.NumAbteilungen*int(r.Rennabstand)) * time.Minute
			curStartTimeSo = curStartTimeSo.Add(rennenDur)

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					pausenDur := time.Duration(p.Laenge) * time.Minute
					curStartTimeSo = curStartTimeSo.Add(pausenDur)

					curMinuteStr := fmt.Sprint(curStartTimeSo.Minute())
					curMinute, err := strconv.Atoi(curMinuteStr[1:])
					if err != nil {
						log.Error(err)
						continue
					}

					roundingMinutes := 10 - curMinute
					roundingDur := time.Duration(roundingMinutes) * time.Minute
					curStartTimeSo = curStartTimeSo.Add(roundingDur)
				}
			}
		} else {
			log.Errorf("RennenNummer %s Tag Error %s", r.Nummer, r.Tag)
		}
	}

	return api.JSON(c, "Zeitplan erfolgreich erstellt!")
}
