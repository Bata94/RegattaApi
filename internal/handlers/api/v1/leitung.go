package api_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetPdfFooter(c *handler.Context) error {
	return c.JSON("footer placeholder")
}

func GetMeldeergebnisList(c *handler.Context) error {
	files, err := utils.GetFilenames("meldeergebnis")
	if err != nil {
		return err
	}
	return c.JSON(files)
}

func GetMeldeergebnisFilename(c *handler.Context) error {
	filename := c.Param("filename")
	filePath := filepath.Join("./files", "meldeergebnis", filename)
	http.ServeFile(c.Writer, c.Request, filePath)
	return nil
}

func GetMeldeergebnisHtml(c *handler.Context) error {
	pLs, err := crud.GetAllPausen()
	if err != nil {
		return err
	}
	rLs, err := crud.GetAllRennenWithAthlet(crud.GetAllRennenParams{
		GetMeldungen:  true,
		GetAthleten:   true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	pLsParsed := []PausesenMeldeergebnisPDF{}
	for _, p := range pLs {
		pLsParsed = append(pLsParsed, PausesenMeldeergebnisPDF{
			Id:             int(p.ID),
			Laenge:         int(p.Laenge),
			NachRennenUuid: p.NachRennenUuid.String(),
		})
	}

	rLsParsed := []RennenMeldeergebnisPDF{}
	for _, r := range rLs {
		rParsed := RennenMeldeergebnisPDF{
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
			Abteilungen:       make([]AbteilungenMeldeergebnisPDF, *r.NumAbteilungen),
			Abmeldungen:       []MeldungMeldeergebnisPDF{},
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

			meldungEntry := MeldungMeldeergebnisPDF{
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
			rParsed.Abteilungen[abteilung-1].Meldungen = append(rParsed.Abteilungen[abteilung-1].Meldungen, mParsed)
		}

		rLsParsed = append(rLsParsed, rParsed)
	}

	return c.JSON(rLsParsed)
}

func GenerateMeldeergebnis(c *handler.Context) error {
	fp, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile(c.Writer, c.Request, fp)
	return nil
}

func DrvMeldungUpload(c *handler.Context) error {
	filename, content, err := c.FormFile("file")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	fmt.Println("Uploaded:", filename)
	uploadsDir := "./tmp/uploads/"

	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		return &handler.Error{StatusCode: 500, Message: "Error while creating uploads directory"}
	}

	dest := fmt.Sprintf("%s%s_%s.json", uploadsDir, "DrvMeldung", time.Now().Format("2006-01-02_15-04-05"))
	err = c.SaveFile(dest, content)
	if err != nil {
		return &handler.Error{StatusCode: 500, Message: err.Error()}
	}

	err = ImportDrvJson(dest)
	if err != nil {
		return &handler.Error{StatusCode: 500, Message: "An Error occurred while importing the JSON File! If you directly downloaded the File from DRV and uploaded it, without modifying it, please contact the Admin! Details: " + err.Error()}
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		return nil
	} else {
		return c.JSON("File uploaded successfully!")
	}
}

func GenerateErgebnisHtml(c *handler.Context) error {
	rLsParsed := []ErgebnisRennenPDF{}
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
			rParsed := ErgebnisRennenPDF{
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
				Abteilungen:       make([]ErgebnisAbteilungPDF, *r.NumAbteilungen),
				Dns:               []MeldungMeldeergebnisPDF{},
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
					rParsed.Dns = append(rParsed.Dns, MeldungMeldeergebnisPDF{
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

				endZeitStr := fmt.Sprintf("%02d:%02d.%03d\n", minutes, secondsPart, milliseconds)

				meldungEntry := ErgebnisMeldungPDF{
					StartNummer: int(m.StartNummer),
					Bahn:        int(m.Bahn),
					Teilnehmer:  athletenStr,
					Verein:      m.Verein.Name,
					Platz:       1,
					Endzeit:     ergebnis.Endzeit,
					EndzeitStr:  endZeitStr,
				}

				abteilung := int(m.Abteilung)
				mParsed := meldungEntry
				rParsed.Abteilungen[abteilung-1].Meldungen = append(rParsed.Abteilungen[abteilung-1].Meldungen, mParsed)
			}

			for i, abt := range rParsed.Abteilungen {
				sort.Slice(abt.Meldungen, func(i, j int) bool {
					return abt.Meldungen[i].Endzeit < abt.Meldungen[j].Endzeit
				})

				p := 1
				for j := range abt.Meldungen {
					rParsed.Abteilungen[i].Meldungen[j].Platz = p
					p++
				}
			}
			rLsParsed = append(rLsParsed, rParsed)
		}
	}

	return c.JSON(rLsParsed)
}

func GenerateErgebnis(c *handler.Context) error {
	fp, err := utils.SavePDFfromHTML(
		"leitung/ergebnis",
		"ergebnis",
		fmt.Sprintf("ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile(c.Writer, c.Request, fp)
	return nil
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

func ImportDrvJson(filePath string) error {
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
	os.WriteFile(fmt.Sprintf("./tmp/Imported_DrvMeldung_%s.json", time.Now().Format("15-04-05")), o, 0o666)

	var apiReqError *api.ReqError
	// TODO: Use a Transaction here!

	for _, v := range drvMeldung.Clubs {
		verein, err := crud.GetVereinMinimal(v.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				fmt.Println("Crud get Error:", err)
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
		_, err = crud.CreateVerein(newVerein)
		if err != nil {
			fmt.Println("Crud create Error:", err)
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
	if err != nil {
		return err
	}

	for _, r := range drvMeldung.Events {
		rennen, err := crud.GetRennenMinimal(r.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				fmt.Println("Crud get Error:", err)
				return err
			}
			if rennen.Uuid != uuid.Nil {
				fmt.Printf("Rennen already exists: %s - %s\n", rennen.Nummer, rennen.BezeichnungLang)
				continue
			}
		}
		if rennen.Uuid != uuid.Nil {
			fmt.Printf("Rennen already exists: %s - %s\n", rennen.Nummer, rennen.BezeichnungLang)
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
			fmt.Println("SortOrder Sonntag!")
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
			fmt.Println("Crud create Error:", err)
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
	fmt.Println("Len All Rennen:", len(allRennen))

	for _, a := range drvMeldung.ClubMembers {
		athlet, err := crud.GetAthletMinimal(a.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				fmt.Println("Crud get Error:", err)
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

		_, err = crud.CreateAthlet(newAthlet)
		if err != nil {
			fmt.Println("Crud create Error:", err)
			return err
		}
	}

	fmt.Println("Import Entries Loop...")
	for _, m := range drvMeldung.Entries {
		meldung, err := crud.GetMeldungMinimal(m.Id)
		if err != nil {
			if !errors.As(err, &apiReqError) {
				fmt.Println("Crud get Error:", err)
				return err
			}
			if meldung.Uuid != uuid.Nil {
				if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
					fmt.Println("Meldung exists in DB, skipping...")
					continue
				}

				fmt.Println("MeldUuid:", meldung.Uuid)
				fmt.Println("Meld in DB Rev:", meldung.DrvRevisionUuid.ClockSequence())
				fmt.Println("Meld in JSON Rev:", m.RevisionId.ClockSequence())

				if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
					fmt.Printf("Meldung in DB is newer than in JSON! MeldungID: %s\n", m.Id)
					continue
				}

				fmt.Printf("Min. eine Meldung in JSON is newer than in DB! MeldungID: %s\n", m.Id)
				continue
			}
		}
		if meldung.Uuid != uuid.Nil {
			if meldung.DrvRevisionUuid.ClockSequence() == m.RevisionId.ClockSequence() {
				fmt.Println("Meldung exists in DB, skipping...")
				continue
			}

			fmt.Println("MeldUuid:", meldung.Uuid)
			fmt.Println("Meld in DB Rev:", meldung.DrvRevisionUuid.ClockSequence())
			fmt.Println("Meld in JSON Rev:", m.RevisionId.ClockSequence())

			if meldung.DrvRevisionUuid.ClockSequence() > m.RevisionId.ClockSequence() {
				fmt.Printf("Meldung in DB is newer than in JSON! MeldungID: %s\n", m.Id)
				continue
			}

			fmt.Printf("Min. eine Meldung in JSON is newer than in DB! MeldungID: %s\n", m.Id)
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
			fmt.Println("Alternativ Meldung gefunden! Status:", m.Status)
			typ += fmt.Sprintf(" - Alternative zu RennenUUID: %s", m.AltEventID.String())
			abgemeldet = true
			kosten = int32(0)
		}

		for _, a := range m.Members {
			role := strings.ToLower(a.Role)
			aUuid := a.ClubMemberId
			if aUuid == uuid.Nil {
				fmt.Println("uuid is nil", aUuid)
				for _, nnA := range allNNAthleten {
					if nnA.VereinUuid == m.ClubId {
						aUuid = nnA.Uuid
						break
					}
				}
				if aUuid == uuid.Nil {
					fmt.Println("uuid is still nil", aUuid)
					return &handler.Error{StatusCode: 500, Message: "Failed to find athlete UUID"}
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
				fmt.Println("Unknown Role:", a.Role)
				continue
			}
			athleten = append(athleten, crud.CreateMeldungAthletParams{
				Uuid:     aUuid,
				Position: int32(a.Position),
				Rolle:    rolle,
			})
		}

		fmt.Println("Members done... Creating Meldung")
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
			fmt.Println("Crud create Error", err)
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
		fmt.Println(m.EventId)
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

func shuffle(array []crud.Meldung) []crud.Meldung {
	for i := range array {
		j := rand.IntN(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func SetzungsLosung(c *handler.Context) error {
	fmt.Println("Setzungslosung!")
	check, err := crud.CheckMeldungSetzung()
	fmt.Println("Check:", check)
	fmt.Println("Err:", err)
	if err != nil {
		return err
	}
	if check {
		return &handler.Error{StatusCode: 400, Message: "Setzung already done!"}
	}

	allRennen, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		fmt.Println("GetAllRennen Error:", err)
		return err
	}

	for _, r := range allRennen {
		shuffledMeldungen := shuffle(r.Meldungen)
		for i, m := range shuffledMeldungen {
			m.Abteilung = int32((i / 6) + 1)
			m.Bahn = int32((i % 6) + 1)
			err = crud.UpdateMeldungSetzung(sqlc.UpdateMeldungSetzungParams{
				Uuid:      m.Uuid,
				Abteilung: m.Abteilung,
				Bahn:      m.Bahn,
			})
			if err != nil {
				fmt.Println("UpdateMeldungSetzung Error:", err)
				return err
			}
		}
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		return nil
	} else {
		return c.JSON("Setzung erfolgreich!")
	}
}

func ResetSetzung(c *handler.Context) error {
	mLs, err := crud.GetAllMeldungen()
	if err != nil {
		return err
	}

	for _, m := range mLs {
		err = crud.UpdateMeldungSetzung(sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: 0,
			Bahn:      0,
		})
		if err != nil {
			return err
		}
	}

	if c.Request.Header.Get("HX-Request") == "true" {
		return nil
	} else {
		return c.JSON("Losung erfolgreich!")
	}
}

func SetStartnummern(c *handler.Context) error {
	check, err := crud.CheckMeldungSetzung()
	if err != nil {
		return err
	}
	if !check {
		return &handler.Error{StatusCode: 400, Message: "Setzung not done!"}
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
		startNummer := int32(1)
		for _, m := range r.Meldungen {
			if m.Abgemeldet {
				continue
			}
			err = crud.UpdateStartNummer(sqlc.UpdateStartNummerParams{
				Uuid:        m.Uuid,
				StartNummer: startNummer,
			})
			if err != nil {
				return err
			}
			startNummer++
		}
	}

	return c.JSON("Startnummern vergeben!")
}

func SetZeitplan(c *handler.Context) error {
	param := new(crud.SetZeitplanParams)
	err := c.BodyParser(param)
	if err != nil {
		return err
	}

	err = crud.SetZeitplan(*param)
	if err != nil {
		return err
	}

	return c.JSON("Zeitplan gesetzt!")
}

type PausesenMeldeergebnisPDF struct {
	Id             int
	Laenge         int
	NachRennenUuid string
}

type RennenMeldeergebnisPDF struct {
	Uuid              string
	RennNr            string
	Bezeichnung       string
	BezeichnungZusatz string
	Startzeit         string
	Rennabstand       int
	Tag               string
	NumMeldungen      int
	NumAbteilungen    int
	Wettkampf         sqlc.Wettkampf
	Abteilungen       []AbteilungenMeldeergebnisPDF
	Abmeldungen       []MeldungMeldeergebnisPDF
}

type AbteilungenMeldeergebnisPDF struct {
	Nummer    int
	Meldungen []MeldungMeldeergebnisPDF
}

type MeldungMeldeergebnisPDF struct {
	StartNummer int
	Bahn        int
	Teilnehmer  string
	Verein      string
}

type ErgebnisRennenPDF struct {
	Uuid              string
	RennNr            string
	Bezeichnung       string
	BezeichnungZusatz string
	Startzeit         string
	Rennabstand       int
	Tag               string
	NumMeldungen      int
	NumAbteilungen    int
	Wettkampf         sqlc.Wettkampf
	Abteilungen       []ErgebnisAbteilungPDF
	Dns               []MeldungMeldeergebnisPDF
}

type ErgebnisAbteilungPDF struct {
	Nummer    int
	Meldungen []ErgebnisMeldungPDF
}

type ErgebnisMeldungPDF struct {
	StartNummer int
	Bahn        int
	Teilnehmer  string
	Verein      string
	Platz       int
	Endzeit     float64
	EndzeitStr  string
}
