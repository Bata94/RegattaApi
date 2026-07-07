package crud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
)

type Meldung struct {
	sqlc.Meldung
	Rennen   *Rennen  `json:"rennen,omitempty"`
	Verein   *Verein  `json:"verein,omitempty"`
	Athleten []Athlet `json:"athleten,omitempty"`
}

func MeldungFromSqlc(m sqlc.Meldung) Meldung {
	return Meldung{Meldung: m}
}

func (m *Meldung) GetAthleten() ([]Athlet, error) {
	if m.Athleten != nil {
		return m.Athleten, nil
	}
	return nil, nil
}

func (m *Meldung) BemerkungStr() string {
	if m.Bemerkung.Valid {
		return m.Bemerkung.String
	}
	return ""
}

func (m *Meldung) ZeitnahmeBemerkungStr() string {
	if m.ZeitnahmeBemerkung.Valid {
		return m.ZeitnahmeBemerkung.String
	}
	return ""
}

func (m *Meldung) RechnungsNummerStr() string {
	if m.RechnungsNummer.Valid {
		return m.RechnungsNummer.String
	}
	return ""
}

type meldungJSON struct {
	Uuid               uuid.UUID `json:"uuid"`
	DrvRevisionUuid    uuid.UUID `json:"drv_revision_uuid"`
	Typ                string    `json:"typ"`
	Bemerkung          *string   `json:"bemerkung"`
	Abgemeldet         bool      `json:"abgemeldet"`
	Dns                bool      `json:"dns"`
	Dnf                bool      `json:"dnf"`
	Dsq                bool      `json:"dsq"`
	ZeitnahmeBemerkung *string   `json:"zeitnahme_bemerkung"`
	StartNummer        int32     `json:"start_nummer"`
	Abteilung          int32     `json:"abteilung"`
	Bahn               int32     `json:"bahn"`
	Kosten             int32     `json:"kosten"`
	RechnungsNummer    *string   `json:"rechnungs_nummer"`
	VereinUuid         uuid.UUID `json:"verein_uuid"`
	RennenUuid         uuid.UUID `json:"rennen_uuid"`
	Rennen             *Rennen   `json:"rennen,omitempty"`
	Verein             *Verein   `json:"verein,omitempty"`
	Athleten           []Athlet  `json:"athleten,omitempty"`
}

func (m Meldung) MarshalJSON() ([]byte, error) {
	j := meldungJSON{
		Uuid:            m.Uuid,
		DrvRevisionUuid: m.DrvRevisionUuid,
		Typ:             m.Typ,
		Abgemeldet:      m.Abgemeldet,
		Dns:             m.Dns,
		Dnf:             m.Dnf,
		Dsq:             m.Dsq,
		StartNummer:     m.StartNummer,
		Abteilung:       m.Abteilung,
		Bahn:            m.Bahn,
		Kosten:          m.Kosten,
		VereinUuid:      m.VereinUuid,
		RennenUuid:      m.RennenUuid,
		Rennen:          m.Rennen,
		Verein:          m.Verein,
	}
	if m.Bemerkung.Valid {
		j.Bemerkung = &m.Bemerkung.String
	}
	if m.ZeitnahmeBemerkung.Valid {
		j.ZeitnahmeBemerkung = &m.ZeitnahmeBemerkung.String
	}
	if m.RechnungsNummer.Valid {
		j.RechnungsNummer = &m.RechnungsNummer.String
	}
	if m.Athleten != nil {
		j.Athleten = m.Athleten
	}
	return json.Marshal(j)
}

func (m Meldung) TeilnehmerString() string {
	var retStr string
	for _, a := range m.Athleten {
		if *a.Rolle == sqlc.RolleTrainer {
			continue
		}

		if retStr != "" {
			retStr += ", "
		}

		if *a.Rolle == sqlc.RolleStm {
			retStr += "Stm.: "
			retStr += a.AthletString()
		} else {
			retStr += a.AthletString()
		}
	}
	return retStr
}

type UpdateSetzungBatchParams struct {
	RennenUUID uuid.UUID                         `json:"rennen_uuid"`
	Meldungen  []sqlc.UpdateMeldungSetzungParams `json:"meldungen"`
}

type CreateMeldungParams struct {
	sqlc.CreateMeldungParams
	Athleten []CreateMeldungAthletParams
}

type CreateMeldungAthletParams struct {
	Uuid     uuid.UUID  `json:"uuid"`
	Position int32      `json:"position"`
	Rolle    sqlc.Rolle `json:"rolle"`
}

func GetAllMeldungen(ctx context.Context) ([]Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	mLs := []Meldung{}
	q, err := DB.Queries.GetAllMeldung(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range q {
		mLs = append(mLs, Meldung{
			Meldung: m,
		})
	}

	return mLs, nil
}

func GetMeldungMinimal(ctx context.Context, uuid uuid.UUID) (Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	m, err := DB.Queries.GetMeldungMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Meldung{}, apierr.ErrNotFound
		}
		return Meldung{}, err
	}

	return Meldung{Meldung: m}, nil
}

func GetMeldungByStartNrUndTag(ctx context.Context, startNummer int, tag Tag) (Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.Queries.GetMeldungByStartNrUndTag(ctx, sqlc.GetMeldungByStartNrUndTagParams{
		StartNummer: int32(startNummer),
		Tag:         sqlc.Tag(tag),
	})
	if err != nil {
		return Meldung{}, err
	}

	if len(q) > 1 {
		return Meldung{}, errors.New("Multiple Startnummern")
	} else if len(q) == 0 {
		slog.Info("No Meldung found", "startNummer", startNummer)
		return Meldung{}, apierr.ErrNotFound
	}

	return Meldung{Meldung: q[0]}, nil
}

func GetMeldung(ctx context.Context, uuid uuid.UUID) (Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.Queries.GetMeldung(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Meldung{}, apierr.ErrNotFound
		}
		return Meldung{}, err
	}

	if len(q) == 0 {
		return Meldung{}, apierr.ErrNotFound
	}

	athleten := []Athlet{}
	rennen := RennenFromSqlc(q[0].Rennen, 0, 0)

	for _, a := range q {
		pos := int(a.Position)
		athleten = append(athleten, Athlet{
			Athlet:   a.Athlet,
			Rolle:    &a.Rolle,
			Position: &pos,
		})
	}

	return Meldung{
		Meldung:  q[0].Meldung,
		Verein:   &Verein{Verein: q[0].Verein},
		Rennen:   &rennen,
		Athleten: athleten,
	}, nil
}

func CheckMeldungSetzung(ctx context.Context) (bool, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	_, err := DB.Queries.CheckMedlungSetzung(ctx)
	if err != nil {
		if isNoRowError(err) {
			return false, nil
		}
		return true, err
	}

	return true, nil
}

func CheckMeldungStartnummern(ctx context.Context) (bool, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	_, err := DB.Queries.CheckMedlungStartnummern(ctx)
	if err != nil {
		if isNoRowError(err) {
			return false, nil
		}
		return true, err
	}

	return true, nil
}

func CreateMeldung(ctx context.Context, mParams CreateMeldungParams) (Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	// TODO: Implement as Transaction!
	m, err := DB.Queries.CreateMeldung(ctx, mParams.CreateMeldungParams)
	if err != nil {
		return Meldung{}, err
	}

	for _, a := range mParams.Athleten {
		_, err = DB.Queries.CreateLinkMeldungAthlet(ctx, sqlc.CreateLinkMeldungAthletParams{
			MeldungUuid: m.Uuid,
			AthletUuid:  a.Uuid,
			Position:    a.Position,
			Rolle:       a.Rolle,
		})

		if err != nil {
			return Meldung{}, apierr.ErrInternal.WithDetails(
				fmt.Sprintf("Error linking MeldungAthlet: %s \nMeldung-ID: %s \nAthlet-ID: %s",
					err,
					m.Uuid.String(),
					a.Uuid.String(),
				),
			)
		}
	}

	return Meldung{Meldung: m}, nil
}

func UpdateMeldungSetzung(ctx context.Context, p sqlc.UpdateMeldungSetzungParams) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.UpdateMeldungSetzung(ctx, p)
}

func UpdateSetzungBatch(ctx context.Context, p UpdateSetzungBatchParams) error {
	if len(p.Meldungen) == 0 {
		return apierr.ErrBadRequest
	}

	for _, m := range p.Meldungen {
		err := UpdateMeldungSetzung(ctx, sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: m.Abteilung,
			Bahn:      m.Bahn,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateStartNummer(ctx context.Context, p sqlc.UpdateStartNummerParams) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.UpdateStartNummer(ctx, p)
}

func GetAllMeldungForVerein(ctx context.Context, vereinUuid uuid.UUID) ([]Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	meldungen := []Meldung{}

	rows, err := DB.Queries.GetAllMeldungForVerein(ctx, vereinUuid)
	if err != nil {
		return meldungen, err
	}

	for i, r := range rows {
		if i == 0 || rows[i-1].Meldung.Uuid != r.Meldung.Uuid {
			rennen := RennenFromSqlc(r.Rennen, 0, int32(0))
			meldungen = append(meldungen, Meldung{
				Meldung:  r.Meldung,
				Rennen:   &rennen,
				Athleten: []Athlet{},
			})
		}

		curMeldung := &meldungen[len(meldungen)-1]

		position := int(r.Position)
		curMeldung.Athleten = append(curMeldung.Athleten, Athlet{
			Athlet:   r.Athlet,
			Rolle:    &r.Rolle,
			Position: &position,
		})
	}

	return meldungen, nil
}

func Ummeldung(ctx context.Context, p sqlc.UmmeldungParams) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.Ummeldung(ctx, p)
}

func Abmeldung(ctx context.Context, meldUuid uuid.UUID) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.Abmeldung(ctx, meldUuid)
}

func SetMeldungRechnungsNummer(ctx context.Context, meldUuid uuid.UUID, rechnungsNummer string) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.SetMeldungRechnungsNummer(ctx, sqlc.SetMeldungRechnungsNummerParams{
		Uuid: meldUuid,
		RechnungsNummer: pgtype.Text{
			Valid:  true,
			String: rechnungsNummer,
		},
	})
}

func GetStartnummerLast(ctx context.Context, tag Tag) (int32, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	lastStartNr, err := DB.Queries.GetLastStartnummer(ctx, sqlc.Tag(tag))
	if err != nil {
		return 0, err
	}

	retInt, ok := lastStartNr.(int32)
	if !ok {
		return 0, errors.New("Last Startnummer nicht umwandelbar!")
	}

	return retInt, nil
}
