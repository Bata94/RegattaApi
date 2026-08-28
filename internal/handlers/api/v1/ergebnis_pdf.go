package api_v1

import (
	jsonv2 "encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/sqlc"
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func GetPdfFooter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := jsonv2.MarshalWrite(w, "footer placeholder"); err != nil {
		slog.Error("failed to encode PDF footer", "error", err)
	}
}

func GetMeldeergebnisList(w http.ResponseWriter, r *http.Request) {
	files, err := utils.GetFilenames("meldeergebnis")
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, files)
}

func GetMeldeergebnisFilename(w http.ResponseWriter, r *http.Request) {
	filename := webfw.Param(r, "filename")
	filePath := filepath.Join(config.C.Paths.FilesDir, "meldeergebnis", filename)
	http.ServeFile(w, r, filePath)
}

func GetMeldeergebnisHtml(w http.ResponseWriter, r *http.Request) {
	pLs, err := crud.GetAllPausen(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	rLs, err := crud.GetAllRennenWithAthlet(r.Context(), crud.GetAllRennenParams{
		GetMeldungen:  true,
		GetAthleten:   true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
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
	for _, race := range rLs {
		zusatz := ""
		if v := race.GetZusatz(); v != nil {
			zusatz = *v
		}
		startzeit := ""
		if v := race.GetStartzeit(); v != nil {
			startzeit = *v
		}
		rennabstand := 0
		if v := race.GetRennabstand(); v != nil {
			rennabstand = *v
		}

		rParsed := pdf_templates.RennenMeldeergebnisPDF{
			Uuid:              race.Uuid.String(),
			RennNr:            race.Nummer,
			Bezeichnung:       race.Bezeichnung,
			BezeichnungZusatz: zusatz,
			Startzeit:         startzeit,
			Rennabstand:       rennabstand,
			Tag:               string(race.Tag),
			NumMeldungen:      *race.NumMeldungen,
			NumAbteilungen:    *race.NumAbteilungen,
			Wettkampf:         race.Wettkampf,
			Abteilungen:       make([]pdf_templates.AbteilungenMeldeergebnisPDF, *race.NumAbteilungen),
			Abmeldungen:       []pdf_templates.MeldungMeldeergebnisPDF{},
		}

		for i := range rParsed.Abteilungen {
			rParsed.Abteilungen[i].Nummer = i + 1
		}

		meldungen, err := race.GetMeldungen(r.Context())
		if err != nil {
			webfw.APIError(w, webfw.InternalError(err.Error()))
			return
		}
		if len(meldungen) == 0 {
			rLsParsed = append(rLsParsed, rParsed)
			continue
		}
		for _, m := range meldungen {
			meldungEntry := pdf_templates.MeldungMeldeergebnisPDF{
				StartNummer: int(m.StartNummer),
				Bahn:        int(m.Bahn),
				Teilnehmer:  m.TeilnehmerString(),
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

	if err := handlers.RenderPdf(
		w,
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		pdf_templates.MeldeErgebnis(rLsParsed, pLsParsed),
	); err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
	}
}

func GenerateMeldeergebnis(w http.ResponseWriter, r *http.Request) {
	fp, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	http.ServeFile(w, r, fp)
}

func GenerateErgebnisHtml(w http.ResponseWriter, r *http.Request) {
	rLsParsed := []ErgebnisRennenPDF{}
	rennen, err := crud.GetAllRennenWithAthlet(r.Context(), crud.GetAllRennenParams{
		GetMeldungen:  true,
		GetAthleten:   true,
		ShowEmpty:     false,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	for _, race := range rennen {
		if race.Wettkampf != sqlc.WettkampfLangstrecke {
			break
		}
		if *race.NumMeldungen != 0 {
			zusatz := ""
			if v := race.GetZusatz(); v != nil {
				zusatz = *v
			}
			startzeit := ""
			if v := race.GetStartzeit(); v != nil {
				startzeit = *v
			}
			rennabstand := 0
			if v := race.GetRennabstand(); v != nil {
				rennabstand = *v
			}

			rParsed := ErgebnisRennenPDF{
				Uuid:              race.Uuid.String(),
				RennNr:            race.Nummer,
				Bezeichnung:       race.Bezeichnung,
				BezeichnungZusatz: zusatz,
				Startzeit:         startzeit,
				Rennabstand:       rennabstand,
				Tag:               string(race.Tag),
				NumMeldungen:      *race.NumMeldungen,
				NumAbteilungen:    *race.NumAbteilungen,
				Wettkampf:         race.Wettkampf,
				Abteilungen:       make([]ErgebnisAbteilungPDF, *race.NumAbteilungen),
				Dns:               []MeldungMeldeergebnisPDF{},
			}

			for i := range rParsed.Abteilungen {
				rParsed.Abteilungen[i].Nummer = i + 1
			}

			meldungen, err := race.GetMeldungen(r.Context())
			if err != nil {
				webfw.APIError(w, webfw.InternalError(err.Error()))
				return
			}
			for _, m := range meldungen {
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

				ergebnis, err := crud.GetZeitnahmeErgebnisByMeld(r.Context(), m.Uuid)
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

	webfw.JSON(w, r, rLsParsed)
}

func GenerateErgebnis(w http.ResponseWriter, r *http.Request) {
	fp, err := utils.SavePDFfromHTML(
		"leitung/ergebnis",
		"ergebnis",
		fmt.Sprintf("ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	http.ServeFile(w, r, fp)
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
