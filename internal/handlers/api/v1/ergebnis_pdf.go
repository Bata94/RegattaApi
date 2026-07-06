package api_v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/sqlc"
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"
	"github.com/bata94/RegattaApi/internal/utils"
)

func GetPdfFooter(c *handler.Context) error {
	return c.JSON("footer placeholder")
}

func GetMeldeergebnisList(c *handler.Context) error {
	files, err := utils.GetFilenames( "meldeergebnis")
	if err != nil {
		return err
	}
	return c.JSON(files)
}

func GetMeldeergebnisFilename(c *handler.Context) error {
	filename := c.Param( "filename")
	filePath := filepath.Join(config.C.Paths.FilesDir, "meldeergebnis", filename)
	http.ServeFile( c.Writer, c.Request, filePath)
	return nil
}

func GetMeldeergebnisHtml(c *handler.Context) error {
	pLs, err := crud.GetAllPausen(c.Request.Context(), )
	if err != nil {
		return err
	}
	rLs, err := crud.GetAllRennenWithAthlet(c.Request.Context(), crud.GetAllRennenParams{
		GetMeldungen:  true,
		GetAthleten:   true,
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
		zusatz := ""
		if v := r.GetZusatz( ); v != nil {
			zusatz = *v
		}
		startzeit := ""
		if v := r.GetStartzeit( ); v != nil {
			startzeit = *v
		}
		rennabstand := 0
		if v := r.GetRennabstand( ); v != nil {
			rennabstand = *v
		}

		rParsed := pdf_templates.RennenMeldeergebnisPDF{
			Uuid:              r.Uuid.String(),
			RennNr:            r.Nummer,
			Bezeichnung:       r.Bezeichnung,
			BezeichnungZusatz: zusatz,
			Startzeit:         startzeit,
			Rennabstand:       rennabstand,
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

		meldungen, err := r.GetMeldungen(c.Request.Context(), )
		if err != nil {
			return err
		}
		if len(meldungen) == 0 {
			rLsParsed = append(rLsParsed, rParsed)
			continue
		}
		for _, m := range meldungen {
			meldungEntry := pdf_templates.MeldungMeldeergebnisPDF{
				StartNummer: int(m.StartNummer),
				Bahn:        int(m.Bahn),
				Teilnehmer:  m.TeilnehmerString( ),
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

	return handlers.RenderPdf( 
		c,
		fmt.Sprintf( "Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		pdf_templates.MeldeErgebnis(rLsParsed, pLsParsed),
	)
}

func GenerateMeldeergebnis(c *handler.Context) error {
	fp, err := utils.SavePDFfromHTML( 
		"leitung/meldeergebnis",
		"meldeergebnis",
		fmt.Sprintf( "Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile( c.Writer, c.Request, fp)
	return nil
}

func GenerateErgebnisHtml(c *handler.Context) error {
	rLsParsed := []ErgebnisRennenPDF{}
	rennen, err := crud.GetAllRennenWithAthlet(c.Request.Context(), crud.GetAllRennenParams{
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
			zusatz := ""
			if v := r.GetZusatz( ); v != nil {
				zusatz = *v
			}
			startzeit := ""
			if v := r.GetStartzeit( ); v != nil {
				startzeit = *v
			}
			rennabstand := 0
			if v := r.GetRennabstand( ); v != nil {
				rennabstand = *v
			}

			rParsed := ErgebnisRennenPDF{
				Uuid:              r.Uuid.String(),
				RennNr:            r.Nummer,
				Bezeichnung:       r.Bezeichnung,
				BezeichnungZusatz: zusatz,
				Startzeit:         startzeit,
				Rennabstand:       rennabstand,
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

			meldungen, err := r.GetMeldungen(c.Request.Context(), )
			if err != nil {
				return err
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
						athletenStr += fmt.Sprintf( "\nStm.: %s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
					} else {
						athletenStr += fmt.Sprintf( "%s %s (%s)", a.Vorname, a.Name, a.Jahrgang)
					}
				}

				ergebnis, err := crud.GetZeitnahmeErgebnisByMeld(c.Request.Context(), m.Uuid)
				if err != nil {
					rParsed.Dns = append(rParsed.Dns, MeldungMeldeergebnisPDF{
						StartNummer: int(m.StartNummer),
						Bahn:        int(m.Bahn),
						Teilnehmer:  athletenStr,
						Verein:      m.Verein.Name,
					})
					continue
				}

				endZeit := time.Duration( ergebnis.Endzeit * float64(time.Second))
				minutes := int(endZeit / time.Minute)
				secondsPart := int((endZeit % time.Minute) / time.Second)
				milliseconds := int((endZeit % time.Second) / time.Millisecond)

				endZeitStr := fmt.Sprintf( "%02d:%02d.%03d\n", minutes, secondsPart, milliseconds)

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
				sort.Slice( abt.Meldungen, func(i, j int) bool {
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
		fmt.Sprintf( "ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile( c.Writer, c.Request, fp)
	return nil
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
