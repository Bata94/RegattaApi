package components

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

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/service"
	"github.com/bata94/RegattaApi/internal/sqlc"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	profil "github.com/bata94/RegattaApi/internal/templates/pages/profil"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func LoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		fieldErrors := make(map[string]string)
		if username == "" {
			fieldErrors["username"] = "Benutzername erforderlich"
		}
		if password == "" {
			fieldErrors["password"] = "Passwort erforderlich"
		}
		webfw.ErrorWithForm(w, r, ui_pages.Login("", fieldErrors), "Bitte alle Felder ausfüllen")
		return
	}

	u, err := crud.AuthLogin(r.Context(), crud.LoginParams{Username: username, Password: password})
	if err != nil {
		webfw.ErrorWithForm(w, r, ui_pages.Login("", nil), "Benutzername oder Passwort ist falsch")
		return
	}

	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    u.Jwt.Token,
		MaxAge:   72 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	webfw.SetRedirect(w, "/internal")
	w.WriteHeader(http.StatusOK)
}

func ImageComponent(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	src := queryParams.Get("src")
	alt := queryParams.Get("alt")

	if src == "" {
		webfw.ErrorToast(w, r, "Image src is empty")
		return
	}

	imgOpt := ui_components.DefaultImageOptions()

	if w := queryParams.Get("width"); w != "" {
		imgOpt.Width = w
	}
	if h := queryParams.Get("height"); h != "" {
		imgOpt.Height = h
	}
	if q := queryParams.Get("quality"); q != "" {
		qFloat64, err := strconv.ParseFloat(q, 32)
		if err != nil || qFloat64 <= 0.0 || qFloat64 > 100.0 {
			slog.Warn("Image component: quality is not a float32 value or out of range... Setting default value")
		} else {
			imgOpt.Quality = float32(qFloat64)
		}
	}
	if l := queryParams.Get("lossless"); l != "" {
		imgOpt.Lossless, _ = strconv.ParseBool(l)
		if imgOpt.Lossless {
			slog.Warn("Image component: lossless is not a bool value... Setting default value")
			imgOpt.Lossless = false
		}
	}
	if class := queryParams.Get("class"); class != "" {
		imgOpt.ClassImage = class
	}

	webfw.SetCacheControl(w, "public, max-age=31536000, immutable")
	templ.Handler(ui_components.RawImageComponent(src, alt, imgOpt)).ServeHTTP(w, r)
}

func UserEditNew(w http.ResponseWriter, r *http.Request) {
	var u *crud.User
	if webfw.Param(r, "uuid") == "" {
		webfw.ErrorToast(w, r, "User not found")
		return
	} else if webfw.Param(r, "uuid") == "new" {
		u = &crud.User{
			User:      sqlc.User{},
			UserGroup: &sqlc.UsersGroup{},
		}
	} else {
		userUUID, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		u, err = crud.GetUser(r.Context(), userUUID)
		if err != nil {
			webfw.ErrorToast(w, r, "User not found")
			return
		}
	}

	templ.Handler(ui_components.UserEdit(*u, "", nil)).ServeHTTP(w, r)
}

func UserEditNewPost(w http.ResponseWriter, r *http.Request) {
	var (
		u        *crud.User
		userUuid uuid.UUID
		err      error
	)

	uuidStr := webfw.Param(r, "uuid")
	if uuidStr == "new" {
		userUuid, err = uuid.NewV7()
	} else {
		userUuid, err = uuid.Parse(uuidStr)
	}

	username := r.FormValue("username")
	groupUuid, errGroupUuid := uuid.Parse(r.FormValue("user_group_uuid"))
	isNotActive := r.FormValue("is_not_active") == "on"

	if err != nil || errGroupUuid != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	fieldErrors := make(map[string]string)
	if username == "" {
		fieldErrors["username"] = "Benutzername erforderlich"
	}
	if groupUuid == uuid.Nil {
		fieldErrors["user_group_uuid"] = "Nutzergruppe erforderlich"
	}
	if uuidStr == "new" {
		password := r.FormValue("password")
		if password == "" {
			fieldErrors["password"] = "Passwort erforderlich"
		}
	}
	if len(fieldErrors) > 0 {
		u = &crud.User{
			User: sqlc.User{
				Uuid:     userUuid,
				Username: username,
				IsActive: !isNotActive,
			},
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	if uuidStr == "new" {
		u = &crud.User{
			User: sqlc.User{
				Uuid:     userUuid,
				Username: username,
				IsActive: !isNotActive,
			},
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		_, err = crud.CreateUser(r.Context(), crud.CreateUserParams{
			GroupUuid: groupUuid,
			Username:  username,
			Password:  r.FormValue("password"),
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while creating user, Err: "+err.Error())
			return
		}

		webfw.SetRedirect(w, "/internal/admin/users")
		w.WriteHeader(http.StatusOK)
		return
	}

	u, err = crud.GetUser(r.Context(), userUuid)
	if err != nil {
		u = &crud.User{
			User: sqlc.User{
				Uuid:     userUuid,
				Username: username,
				IsActive: !isNotActive,
			},
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while updating user, Err: "+err.Error())
		return
	}

	err = crud.UpdateUser(r.Context(), u.Uuid, crud.UpdateUserParams{
		Username:  username,
		IsActive:  r.FormValue("is_not_active") != "on",
		GroupUuid: groupUuid,
	})
	if err != nil {
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while updating user, Err: "+err.Error())
		return
	}

	webfw.SetRedirect(w, "/internal/admin/users")
	w.WriteHeader(http.StatusOK)
}

func UserGroupEditNew(w http.ResponseWriter, r *http.Request) {
	var ug sqlc.UsersGroup
	if webfw.Param(r, "uuid") == "" {
		webfw.ErrorToast(w, r, "UserGroup not found")
		return
	} else if webfw.Param(r, "uuid") == "new" {
		ug = sqlc.UsersGroup{}
	} else {
		grpUUID, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		ug, err = crud.GetUsersGroupsMinimal(r.Context(), grpUUID)
		if err != nil {
			webfw.ErrorToast(w, r, "UserGroup not found")
			return
		}
	}

	templ.Handler(ui_components.UserGroupEdit(ug, "", nil)).ServeHTTP(w, r)
}

func UserGroupEditNewPost(w http.ResponseWriter, r *http.Request) {
	uuidStr := webfw.Param(r, "uuid")
	var groupUuid uuid.UUID

	if uuidStr == "new" {
		var err error
		groupUuid, err = uuid.NewV7()
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	} else {
		var err error
		groupUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	}

	name := r.FormValue("name")

	fieldErrors := make(map[string]string)
	if name == "" {
		fieldErrors["name"] = "Gruppenname erforderlich"
	}

	capList := []struct {
		formName string
		value    sqlc.UserCapability
	}{
		{"allowed_admin", sqlc.UserCapabilityAllowedAdmin},
		{"allowed_zeitnahme", sqlc.UserCapabilityAllowedZeitnahme},
		{"allowed_startlisten", sqlc.UserCapabilityAllowedStartlisten},
		{"allowed_regattabuero", sqlc.UserCapabilityAllowedRegattabuero},
		{"allowed_regattaleitung", sqlc.UserCapabilityAllowedRegattaleitung},
	}
	var capabilities []sqlc.UserCapability
	for _, cf := range capList {
		if r.FormValue(cf.formName) == "on" {
			capabilities = append(capabilities, cf.value)
		}
	}

	ug := sqlc.UsersGroup{
		Uuid:         groupUuid,
		Name:         name,
		Capabilities: capabilities,
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	if uuidStr == "new" {
		_, err := crud.CreateUserGroup(r.Context(), sqlc.CreateUserGroupParams{
			Name:         name,
			Capabilities: capabilities,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", nil), "Fehler beim Erstellen der Nutzergruppe")
			return
		}
	} else {
		err := crud.UpdateUserGroup(r.Context(), groupUuid, sqlc.UpdateUserGroupParams{
			Name:         name,
			Capabilities: capabilities,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", nil), "Fehler beim Aktualisieren der Nutzergruppe")
			return
		}
	}

	webfw.SetRedirect(w, "/internal/admin/usergroups")
	w.WriteHeader(http.StatusOK)
}

func ChangePasswordGet(w http.ResponseWriter, r *http.Request) {
	userUuidStr := webfw.Param(r, "uuid")
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	user, err := crud.GetUser(r.Context(), userUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "User not found")
		return
	}

	templ.Handler(profil.ChangePasswordDialogBody(*user, "", nil)).ServeHTTP(w, r)
}

func ChangePasswordPost(w http.ResponseWriter, r *http.Request) {
	userUuidStr := webfw.Param(r, "uuid")
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	user, err := crud.GetUser(r.Context(), userUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "User not found")
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword1 := r.FormValue("new_password_1")
	newPassword2 := r.FormValue("new_password_2")

	fieldErrors := make(map[string]string)
	topMsg := ""
	if currentPassword == "" {
		fieldErrors["current_password"] = "Aktuelles Passwort erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if newPassword1 == "" {
		fieldErrors["new_password_1"] = "Neues Passwort erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if newPassword2 == "" {
		fieldErrors["new_password_2"] = "Neues Passwort erneut erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if newPassword1 != "" && newPassword2 != "" && newPassword1 != newPassword2 {
		fieldErrors["new_password_1"] = "Passwörter stimmen nicht überein"
		fieldErrors["new_password_2"] = "Passwörter stimmen nicht überein"
		topMsg = "Passwörter stimmen nicht überein"
	}
	if currentPassword != "" && !crud.CheckPasswordHash(currentPassword, user.HashedPassword) {
		fieldErrors["current_password"] = "Aktuelles Passwort ist falsch"
		topMsg = "Aktuelles Passwort ist falsch"
	}
	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, profil.ChangePasswordDialogBody(*user, "", fieldErrors), topMsg)
		return
	}

	err = crud.UpdatePassword(r.Context(), userUuid, newPassword1)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while updating password")
		return
	}

	webfw.SetRedirect(w, "/internal/profil")
	w.WriteHeader(http.StatusOK)
}

func DrvUploadPost(w http.ResponseWriter, r *http.Request) {
	api_v1.DrvMeldungUpload(w, r)
	webfw.SuccessToast(w, r, "Upload erfolgreich!")
}

func SetzungsVerwaltungLosungPost(w http.ResponseWriter, r *http.Request) {
	api_v1.SetzungsLosung(w, r)
	webfw.SuccessToast(w, r, "Losung erfolgreich!")
}

func SetzungsVerwaltungLosungDelete(w http.ResponseWriter, r *http.Request) {
	api_v1.ResetSetzung(w, r)
	webfw.SuccessToast(w, r, "Setzung erfolgreich zurückgesetzt!")
}

func SetzungsVerwaltungAenderungRennenPost(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		rUuid  uuid.UUID
		rennen crud.Rennen
	)

	rUuid, err = uuid.Parse(webfw.Param(r, "param"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	rennen, err = crud.GetRennen(r.Context(), rUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Rennen nicht gefunden")
		return
	}

	payloadStr := r.FormValue("params")
	payload := make(map[string]any)
	err = json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid JSON")
		return
	}

	meldOrderLs, ok := payload["order"].([]any)
	if !ok {
		webfw.ErrorToast(w, r, "Order nicht gefunden")
		return
	}
	abteilungParam := payload["abteilung"]
	if abteilungParam == nil {
		webfw.ErrorToast(w, r, "Abteilung nicht gefunden")
		return
	}
	targetAbteilung := int32(abteilungParam.(float64))

	for i, m := range meldOrderLs {
		mUuid, err := uuid.Parse(m.(string))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}

		for _, meldung := range rennen.Meldungen {
			if meldung.Uuid == mUuid {
				bahn := int32(i) + 1

				err = crud.UpdateMeldungSetzung(r.Context(), sqlc.UpdateMeldungSetzungParams{
					Uuid:      meldung.Uuid,
					Abteilung: targetAbteilung,
					Bahn:      bahn,
				})
				if err != nil {
					webfw.ErrorToast(w, r, "Error while updating meldung setzung")
					return
				}
				continue
			}
		}
	}

	webfw.SuccessToast(w, r, "Setzung erfolgreich!")
}

func StartnummernAendernPost(w http.ResponseWriter, r *http.Request) {
	rennenUuidStr := webfw.Param(r, "r_uuid")
	meldungUuidStr := webfw.Param(r, "m_uuid")

	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	m, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Meldung nicht gefunden")
		return
	}
	if m.RennenUuid != rennenUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	fieldErrors := make(map[string]string)
	startnummer := r.FormValue("startnummer")
	if startnummer == "" {
		fieldErrors["startnummer"] = "Startnummer erforderlich"
	}
	startNummerInt, err := strconv.Atoi(startnummer)
	if err != nil {
		fieldErrors["startnummer"] = "Ungültige Startnummer"
	}

	bereich, err := crud.GetStartnummernBereich(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading startnummernbereich")
		return
	}

	startNummerInt32 := int32(startNummerInt)
	if !bereich.InBereich(startNummerInt32) {
		fieldErrors["startnummer"] = fmt.Sprintf("Startnummer muss zwischen %d und %d liegen", bereich.MinNummer, bereich.MaxNummer)
	}
	if startNummerInt32 > 0 && bereich.IsFehlend(startNummerInt32) {
		fieldErrors["startnummer"] = "Startnummer ist als fehlend markiert"
	}

	checkStartnummer, err := crud.GetMeldungByStartNrUndTag(r.Context(), startNummerInt, m.Rennen.Tag)
	if err != nil && !errors.As(err, &apierr.ErrNotFound) {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}
	if checkStartnummer.Uuid != uuid.Nil {
		fieldErrors["startnummer"] = "Startnummer bereits vergeben"
	}

	if len(fieldErrors) > 0 {
		templ.Handler(regattaleitung.StartnummernAendern(m, fieldErrors)).ServeHTTP(w, r)
		return
	}

	err = crud.UpdateStartNummer(r.Context(), sqlc.UpdateStartNummerParams{
		Uuid:        m.Uuid,
		StartNummer: int32(startNummerInt),
	})
	if err != nil {
		slog.Error("UpdateStartNummer error", "err", err)
		webfw.ErrorToast(w, r, "Error while updating startnummer")
		return
	}

	m, err = crud.GetMeldung(r.Context(), m.Uuid)
	if err != nil {
		slog.Error("GetMeldung error", "err", err)
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	templ.Handler(regattaleitung.StartnummernAendern(m, fieldErrors)).ServeHTTP(w, r)
}

func StartnummernBereichPost(w http.ResponseWriter, r *http.Request) {
	fieldErrors := make(map[string]string)

	minStr := r.FormValue("min_nummer")
	maxStr := r.FormValue("max_nummer")

	minNummer, err := strconv.Atoi(minStr)
	if err != nil || minNummer < 1 {
		fieldErrors["min_nummer"] = "Kleinste Startnummer muss mindestens 1 sein"
	}
	maxNummer, err := strconv.Atoi(maxStr)
	if err != nil || maxNummer < 1 {
		fieldErrors["max_nummer"] = "Größte Startnummer muss mindestens 1 sein"
	}

	fehlendeStr := r.FormValue("fehlende_nummern")
	fehlende, err := parseFehlendeNummern(fehlendeStr)
	if err != nil {
		fieldErrors["fehlende_nummern"] = "Ungültige fehlende Startnummern"
	}

	if minNummer >= 1 && maxNummer >= 1 && maxNummer < minNummer {
		fieldErrors["max_nummer"] = "Größte Startnummer muss größer oder gleich der kleinsten sein"
	}

	if minNummer >= 1 && maxNummer >= minNummer {
		for _, n := range fehlende {
			if n < int32(minNummer) || n > int32(maxNummer) {
				fieldErrors["fehlende_nummern"] = "Fehlende Startnummern müssen im Bereich liegen"
				break
			}
		}
	}

	b := crud.StartnummernBereichFromSqlc(sqlc.StartnummernBereich{
		MinNummer:       int32(minNummer),
		MaxNummer:       int32(maxNummer),
		FehlendeNummern: fehlende,
	})

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.StartnummernBereich(b, fieldErrors), "Ungültiger Startnummernbereich")
		return
	}

	if _, err := crud.SetStartnummernBereich(r.Context(), int32(minNummer), int32(maxNummer), fehlende); err != nil {
		slog.Error("SetStartnummernBereich error", "err", err)
		webfw.ErrorToast(w, r, "Error while saving startnummernbereich")
		return
	}

	webfw.SuccessToast(w, r, "Startnummernbereich gespeichert")
}

func parseFehlendeNummern(s string) ([]int32, error) {
	if strings.TrimSpace(s) == "" {
		return []int32{}, nil
	}

	parts := strings.Split(s, ",")
	seen := make(map[int32]struct{})
	ret := make([]int32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 {
			return nil, errors.New("invalid number")
		}
		if _, ok := seen[int32(n)]; ok {
			continue
		}
		seen[int32(n)] = struct{}{}
		ret = append(ret, int32(n))
	}
	return ret, nil
}

func PausenNew(w http.ResponseWriter, r *http.Request) {
	nachRennenUuidStr := webfw.Param(r, "nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	p := crud.Pause{Pause: sqlc.Pause{ID: int32(-2), NachRennenUuid: nachRennenUuid, Laenge: 45}}

	templ.Handler(regattaleitung.PausenEntry(p)).ServeHTTP(w, r)
}

func PausenPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid ID")
		return
	}
	laengeStr := r.FormValue("laenge")
	laenge, err := strconv.Atoi(laengeStr)
	if err != nil || laenge < 0 || laenge > 120 {
		slog.Error(fmt.Sprintf("Laenge: %s - Error: %s", laengeStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid laenge")
		return
	}
	nachRennenUuidStr := r.FormValue("nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		slog.Error(fmt.Sprintf("UUID: %s - Error: %s", nachRennenUuidStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	if id == -2 {
		_, err = crud.CreatePause(r.Context(), sqlc.CreatePauseParams{
			NachRennenUuid: nachRennenUuid,
			Laenge:         int32(laenge),
		})
		if err != nil {
			webfw.ErrorToast(w, r, "Error while creating pause")
			return
		}

		templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
	} else {
		_, err = crud.UpdatePause(r.Context(), sqlc.UpdatePauseParams{
			ID:     int32(id),
			Laenge: int32(laenge),
		})
		if err != nil {
			webfw.ErrorToast(w, r, "Error while updating pause")
			return
		}

		templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
	}
}

func PausenDelete(w http.ResponseWriter, r *http.Request) {
	idStr := webfw.Param(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid ID")
		return
	}

	err = crud.DeletePause(r.Context(), int32(id))
	if err != nil {
		webfw.ErrorToast(w, r, "Error while deleting pause")
		return
	}

	templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
}

func ZeitplanPost(w http.ResponseWriter, r *http.Request) {
	startzeit_saStr := r.FormValue("startzeit_sa")
	startzeit_soStr := r.FormValue("startzeit_so")

	fieldErrors := make(map[string]string)

	startzeit_sa, err := strconv.Atoi(startzeit_saStr)
	if err != nil || startzeit_sa < 0 || startzeit_sa > 24 {
		fieldErrors["startzeit_sa"] = "Ungültige Startzeit (0-24)"
	}
	startzeit_so, err := strconv.Atoi(startzeit_soStr)
	if err != nil || startzeit_so < 0 || startzeit_so > 24 {
		fieldErrors["startzeit_so"] = "Ungültige Startzeit (0-24)"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.Zeitplan("", fieldErrors), "Ungültige Startzeit")
		return
	}

	zeitplan := service.SetZeitplanParams{
		SaStartStunde: startzeit_sa,
		SoStartStunde: startzeit_so,
	}

	err = service.SetZeitplan(r.Context(), zeitplan)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while creating zeitplan")
		return
	}

	webfw.SuccessToast(w, r, "Zeitplan erstellt")
}

func StartnummernVerteilenPost(w http.ResponseWriter, r *http.Request) {
	err := service.SetStartnummern(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, fmt.Sprintf("Error while setting startnummern: %s", err.Error()))
		return
	}

	webfw.SuccessToast(w, r, "Startnummern erfolgreich verteilt!")
}

func StartnummernVerteilenDelete(w http.ResponseWriter, r *http.Request) {
	err := service.ResetStartnummern(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, "Error while resetting startnummern")
		return
	}

	webfw.SuccessToast(w, r, "Startnummern erfolgreich zurückgesetzt!")
}

func PdfMeldeergebnisPost(w http.ResponseWriter, r *http.Request) {
	fileName := fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05"))
	_, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fileName,
		true,
	)
	if err != nil {
		if rmErr := os.Remove(fmt.Sprintf("%smeldeergebnis/%s", config.C.Paths.FilesDir, fileName)); rmErr != nil {
			slog.Error("Error removing failed PDF file", "err", rmErr)
		}
		webfw.ErrorToast(w, r, fmt.Sprintf("Fehler während PDF Erstellung: %s", err.Error()))
		return
	}

	templ.Handler(regattaleitung.PdfMeldeergebnis(true)).ServeHTTP(w, r)
}

func AbmeldungDelete(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	_, err = crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	if meldung.VereinUuid != vereinUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	err = crud.Abmeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while deleting meldung")
		return
	}

	webfw.SetRedirect(w, fmt.Sprintf("/internal/regattabuero/%s/abmeldung", vereinUuid))
	w.WriteHeader(http.StatusOK)
}

func UmmeldungPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	if err := r.ParseForm(); err != nil {
		webfw.ErrorToast(w, r, "Error parsing form")
		return
	}

	fieldErrors := make(map[string]string)
	for i := range meldung.Athleten {
		athUuidStr := r.FormValue(fmt.Sprintf("athleten_%d", i))
		if athUuidStr == "" {
			continue
		}
		athUuid, err := uuid.Parse(athUuidStr)
		if err != nil {
			fieldErrors[fmt.Sprintf("athleten_%d", i)] = "Ungültige UUID"
			continue
		}
		if athUuid == meldung.Athleten[i].Uuid {
			continue
		}
		err = crud.Ummeldung(r.Context(), sqlc.UmmeldungParams{
			MeldungUuid: meldungUuid,
			Rolle:       *meldung.Athleten[i].Rolle,
			Position:    int32(*meldung.Athleten[i].Position),
			AthletUuid:  athUuid,
		})
		if err != nil {
			fieldErrors[fmt.Sprintf("athleten_%d", i)] = "Fehler beim Aktualisieren"
		}
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.UmmeldungMeldung(verein, meldung, athleten, "", fieldErrors), "Fehler bei der Ummeldung")
		return
	}

	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), verein.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldungen")
		return
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/ummeldung", verein.Uuid))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.Ummeldung(verein, meldungen).Render(context.Background(), w); err != nil {
		slog.Warn("Ummeldung render error", "err", err)
	}
}

func NachmeldungPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	rennenUuidStr := webfw.Param(r, "r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	rennen, err := crud.GetRennen(r.Context(), rennenUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading rennen")
		return
	}

	if err := r.ParseForm(); err != nil {
		webfw.ErrorToast(w, r, "Error parsing form")
		return
	}

	vereinUuid, err := uuid.Parse(r.FormValue("verein_uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	numAthletes, stmRequired := rennen.GetTeilnehmerMeldeParams()

	params := api_v1.PostNachmeldungParams{
		VereinUuid:                    r.FormValue("verein_uuid"),
		RennenUuid:                    r.FormValue("rennen_uuid"),
		DoppeltesMeldentgeldBefreiung: r.FormValue("doppeltes_meldentgeld_befreiung") != "",
		Athleten:                      []api_v1.PostNachmeldungAthletParams{},
	}

	fieldErrors := make(map[string]string)
	hasAthlete := false
	for i := range numAthletes {
		athVal := r.FormValue(fmt.Sprintf("athleten_%d", i))
		if athVal == "" || athVal == "---" {
			continue
		}
		hasAthlete = true
		params.Athleten = append(params.Athleten, api_v1.PostNachmeldungAthletParams{
			AthletUuid: athVal,
			Position:   strconv.Itoa(i),
		})
	}

	if stmRequired {
		stmVal := r.FormValue("stm")
		if stmVal != "" && stmVal != "---" {
			hasAthlete = true
			params.Athleten = append(params.Athleten, api_v1.PostNachmeldungAthletParams{
				AthletUuid: stmVal,
				Position:   "stm",
			})
		}
	}

	if !hasAthlete {
		fieldErrors["athleten_0"] = "Mindestens ein Teilnehmer erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.NachmeldungMeldung(verein, rennen, athleten, "", fieldErrors), "Bitte wähle mindestens einen Teilnehmer aus")
		return
	}

	m, err := api_v1.CreateNachmeldung(r.Context(), params)
	if err != nil {
		webfw.ErrorToast(w, r, "Error creating nachmeldung: "+err.Error())
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), m.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/nachmeldung/success/%s", vereinUuidStr, m.Uuid.String()))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.NachmeldungSuccess(meldung).Render(context.Background(), w); err != nil {
		slog.Warn("NachmeldungSuccess render error", "err", err)
	}
}

func RennenTab(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	showEmpty := webfw.Query(r, "show_empty") == "true"
	showStarted := webfw.Query(r, "show_started") == "true"
	urlFormatStr := webfw.Query(r, "url_format_str")

	templ.Handler(ui_components.RennenTab(wettkampf, urlFormatStr, showEmpty, showStarted)).ServeHTTP(w, r)
}

func NewAthletPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}

	vorname := r.FormValue("vorname")
	name := r.FormValue("name")
	jahrgang := r.FormValue("jahrgang")
	geschlecht := r.FormValue("geschlecht")
	startberechtigt := r.FormValue("startberechtigt") == "on"

	fieldErrors := make(map[string]string)
	if vorname == "" {
		fieldErrors["vorname"] = "Vorname erforderlich"
	}
	if name == "" {
		fieldErrors["name"] = "Name erforderlich"
	}
	if jahrgang == "" {
		fieldErrors["jahrgang"] = "Jahrgang erforderlich"
	}
	if geschlecht != "m" && geschlecht != "w" && geschlecht != "x" {
		fieldErrors["geschlecht"] = "Geschlecht erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.NewAthlet(verein, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	athletUuid, err := uuid.NewV7()
	if err != nil {
		webfw.ErrorToast(w, r, "Error generating UUID")
		return
	}

	a, err := crud.CreateAthlet(r.Context(), sqlc.CreateAthletParams{
		Uuid:            athletUuid,
		VereinUuid:      vereinUuid,
		Name:            name,
		Vorname:         vorname,
		Jahrgang:        jahrgang,
		Startberechtigt: startberechtigt,
		Geschlecht:      sqlc.Geschlecht(geschlecht),
	})
	if err != nil {
		webfw.ErrorWithForm(w, r, regattabuero.NewAthlet(verein, "", nil), "Fehler beim Anlegen des Athleten")
		return
	}

	a.Verein = &verein
	if err := regattabuero.NewAthletSuccess(a).Render(context.Background(), w); err != nil {
		slog.Warn("NewAthletSuccess render error", "err", err)
	}
}

func WaagePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("ParseForm error", "err", err)
		webfw.ErrorToast(w, r, "Fehler beim Verarbeiten der Anfrage")
		return
	}

	idStr := r.FormValue("uuid")
	gewichtStr := r.FormValue("gewicht")

	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Parse UUID error", "err", err)
		webfw.ErrorToast(w, r, "Ungültige UUID")
		return
	}

	ath, err := crud.GetAthletMinimal(r.Context(), id)
	if err != nil {
		slog.Error("GetAthletMinimal error", "err", err)
		webfw.ErrorToast(w, r, err.Error())
		return
	}

	fieldErrors := make(map[string]string)
	gewichtFloat, err := strconv.ParseFloat(gewichtStr, 32)
	if err != nil {
		fieldErrors["gewicht"] = "Ungültiges Gewicht"
	}
	gewicht := int(gewichtFloat * 10)

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.Waage(ath, "", fieldErrors), "Ungültiges Gewicht")
		return
	}

	err = ath.UpdateGewicht(r.Context(), gewicht)
	if err != nil {
		slog.Error("UpdateGewicht error", "err", err)
		webfw.ErrorToast(w, r, "Fehler beim Aktualisieren des Gewichts")
		return
	}

	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Verein nicht gefunden")
		return
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(r.Context(), verein.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/waage", vereinUuidStr))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.WaageWahl(verein, athleten).Render(context.Background(), w); err != nil {
		slog.Warn("WaageWahl render error", "err", err)
	}
}

func StartberechtigungPost(w http.ResponseWriter, r *http.Request) {
	slog.Debug("StartberechtigungPost", "formVal", r.FormValue("startberechtigt"))

	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	if athletUuidStr != r.FormValue("uuid") {
		webfw.ErrorToast(w, r, "UUIDs stimmen nicht überein")
		return
	}
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	_, err = crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athlet")
		return
	}

	if athlet.VereinUuid != vereinUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	formVal := r.FormValue("startberechtigt")
	formVal = strings.ToLower(formVal)
	if formVal != "on" && formVal != "true" {
		webfw.ErrorToast(w, r, "Bitte aktivieren Sie die Ärztliche Bescheinigung")
		return
	}

	err = athlet.UpdateStartberechtigung(r.Context(), true)
	if err != nil {
		slog.Error("UpdateStartberechtigung error", "err", err)
		webfw.ErrorToast(w, r, "Error while updating startberechtigung")
		return
	}

	webfw.SetRedirect(w, fmt.Sprintf("/internal/regattabuero/%s/startberechtigung", vereinUuidStr))
	w.WriteHeader(http.StatusOK)
}

func ZeitplanCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_components.ZeitplanCollapseBody(wettkampf)).ServeHTTP(w, r)
}

func AusschreibungRennenCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_pages.AusschreibungRennenCollapseBody(wettkampf)).ServeHTTP(w, r)
}

func MeldeergebnisCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_pages.MeldeergebnisCollapseBody(wettkampf)).ServeHTTP(w, r)
}
