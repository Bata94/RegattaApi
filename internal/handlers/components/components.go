package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/service"
	"github.com/bata94/RegattaApi/internal/sqlc"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/google/uuid"
)

func LoginPost(c *handler.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		fieldErrors := make(map[string]string)
		if username == "" {
			fieldErrors["username"] = "Benutzername erforderlich"
		}
		if password == "" {
			fieldErrors["password"] = "Passwort erforderlich"
		}
		return handler.BadRequest("Bitte alle Felder ausfüllen").WithForm(ui_pages.Login("", fieldErrors))
	}

	u, err := crud.AuthLogin(c.Request.Context(), crud.LoginParams{Username: username, Password: password})
	if err != nil {
		return handler.BadRequest("Benutzername oder Passwort ist falsch").WithForm(ui_pages.Login("", nil))
	}

	secure := c.Request.TLS != nil
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    u.Jwt.Token,
		MaxAge:   72 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	c.Writer.Header().Set("HX-Redirect", "/")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func ImageComponent(c *handler.Context) error {
	queryParams := c.Request.URL.Query()
	src := queryParams.Get("src")
	alt := queryParams.Get("alt")

	if src == "" {
		return handler.NotFound("Image src is empty")
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

	templ.Handler(ui_components.RawImageComponent(src, alt, imgOpt)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func UserEditNew(c *handler.Context) error {
	var u *crud.User
	if c.Param("uuid") == "" {
		return handler.NotFound("User not found")
	} else if c.Param("uuid") == "new" {
		u = &crud.User{
			User:      sqlc.User{},
			UserGroup: &sqlc.UsersGroup{},
		}
	} else {
		uuid, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			return handler.NotAcceptable("Invalid UUID")
		}
		u, err = crud.GetUser(c.Request.Context(), uuid)
		if err != nil {
			return handler.NotFound("User not found")
		}
	}

	templ.Handler(ui_components.UserEdit(*u, "", nil)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func UserEditNewPost(c *handler.Context) error {
	var (
		u        *crud.User
		userUuid uuid.UUID
		err      error
	)

	uuidStr := c.Param("uuid")
	if uuidStr == "new" {
		userUuid, err = uuid.NewV7()
	} else {
		userUuid, err = uuid.Parse(uuidStr)
	}

	username := c.FormValue("username")
	groupUuid, errGroupUuid := uuid.Parse(c.FormValue("user_group_uuid"))
	isNotActive := c.FormValue("is_not_active") == "on"

	if err != nil || errGroupUuid != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	fieldErrors := make(map[string]string)
	if username == "" {
		fieldErrors["username"] = "Benutzername erforderlich"
	}
	if groupUuid == uuid.Nil {
		fieldErrors["user_group_uuid"] = "Nutzergruppe erforderlich"
	}
	if uuidStr == "new" {
		password := c.FormValue("password")
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
		return handler.BadRequest("Bitte alle Pflichtfelder ausfüllen").WithForm(ui_components.UserEdit(*u, "", fieldErrors))
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
		_, err = crud.CreateUser(c.Request.Context(), crud.CreateUserParams{
			GroupUuid: groupUuid,
			Username:  username,
			Password:  c.FormValue("password"),
		})
		if err != nil {
			return handler.BadRequest("Error while creating user, Err: " + err.Error()).WithForm(ui_components.UserEdit(*u, "", nil))
		}

		c.Writer.Header().Set("HX-Redirect", "/internal/admin/users")
		c.Writer.WriteHeader(http.StatusOK)
		return nil
	}

	u, err = crud.GetUser(c.Request.Context(), userUuid)
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
		return handler.BadRequest("Error while updating user, Err: " + err.Error()).WithForm(ui_components.UserEdit(*u, "", nil))
	}

	err = crud.UpdateUser(c.Request.Context(), u.Uuid, crud.UpdateUserParams{
		Username:  username,
		IsActive:  c.FormValue("is_not_active") != "on",
		GroupUuid: groupUuid,
	})
	if err != nil {
		return handler.BadRequest("Error while updating user, Err: " + err.Error()).WithForm(ui_components.UserEdit(*u, "", nil))
	}

	c.Writer.Header().Set("HX-Redirect", "/internal/admin/users")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func UserGroupEditNew(c *handler.Context) error {
	var ug sqlc.UsersGroup
	if c.Param("uuid") == "" {
		return handler.NotFound("UserGroup not found")
	} else if c.Param("uuid") == "new" {
		ug = sqlc.UsersGroup{}
	} else {
		uuid, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			return handler.NotAcceptable("Invalid UUID")
		}
		ug, err = crud.GetUsersGroupsMinimal(c.Request.Context(), uuid)
		if err != nil {
			return handler.NotFound("UserGroup not found")
		}
	}

	templ.Handler(ui_components.UserGroupEdit(ug, "", nil)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func UserGroupEditNewPost(c *handler.Context) error {
	uuidStr := c.Param("uuid")
	var groupUuid uuid.UUID

	if uuidStr == "new" {
		var err error
		groupUuid, err = uuid.NewV7()
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	} else {
		var err error
		groupUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	}

	name := c.FormValue("name")

	fieldErrors := make(map[string]string)
	if name == "" {
		fieldErrors["name"] = "Gruppenname erforderlich"
	}

	ug := sqlc.UsersGroup{
		Uuid:                  groupUuid,
		Name:                  name,
		AllowedAdmin:          c.FormValue("allowed_admin") == "on",
		AllowedZeitnahme:      c.FormValue("allowed_zeitnahme") == "on",
		AllowedStartlisten:    c.FormValue("allowed_startlisten") == "on",
		AllowedRegattabuero:   c.FormValue("allowed_regattabuero") == "on",
		AllowedRegattaleitung: c.FormValue("allowed_regattaleitung") == "on",
	}

	if len(fieldErrors) > 0 {
		return handler.BadRequest("Bitte alle Pflichtfelder ausfüllen").WithForm(ui_components.UserGroupEdit(ug, "", fieldErrors))
	}

	if uuidStr == "new" {
		_, err := crud.CreateUserGroup(c.Request.Context(), sqlc.CreateUserGroupParams{
			Name:                  name,
			AllowedAdmin:          ug.AllowedAdmin,
			AllowedZeitnahme:      ug.AllowedZeitnahme,
			AllowedStartlisten:    ug.AllowedStartlisten,
			AllowedRegattabuero:   ug.AllowedRegattabuero,
			AllowedRegattaleitung: ug.AllowedRegattaleitung,
		})
		if err != nil {
			return handler.BadRequest("Fehler beim Erstellen der Nutzergruppe").WithForm(ui_components.UserGroupEdit(ug, "", nil))
		}
	} else {
		err := crud.UpdateUserGroup(c.Request.Context(), groupUuid, sqlc.UpdateUserGroupParams{
			Name:                  name,
			AllowedAdmin:          ug.AllowedAdmin,
			AllowedZeitnahme:      ug.AllowedZeitnahme,
			AllowedStartlisten:    ug.AllowedStartlisten,
			AllowedRegattabuero:   ug.AllowedRegattabuero,
			AllowedRegattaleitung: ug.AllowedRegattaleitung,
		})
		if err != nil {
			return handler.BadRequest("Fehler beim Aktualisieren der Nutzergruppe").WithForm(ui_components.UserGroupEdit(ug, "", nil))
		}
	}

	c.Writer.Header().Set("HX-Redirect", "/internal/admin/usergroups")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func ChangePasswordGet(c *handler.Context) error {
	userUuidStr := c.Param("uuid")
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	user, err := crud.GetUser(c.Request.Context(), userUuid)
	if err != nil {
		return handler.NotFound("User not found")
	}

	templ.Handler(ui_pages.ChangePasswordDialogBody(*user, "", nil)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func ChangePasswordPost(c *handler.Context) error {
	userUuidStr := c.Param("uuid")
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	user, err := crud.GetUser(c.Request.Context(), userUuid)
	if err != nil {
		return handler.NotFound("User not found")
	}

	currentPassword := c.FormValue("current_password")
	newPassword1 := c.FormValue("new_password_1")
	newPassword2 := c.FormValue("new_password_2")

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
		return handler.BadRequest(topMsg).WithForm(ui_pages.ChangePasswordDialogBody(*user, "", fieldErrors))
	}

	err = crud.UpdatePassword(c.Request.Context(), userUuid, newPassword1)
	if err != nil {
		return handler.InternalError("Error while updating password")
	}

	c.Writer.Header().Set("HX-Redirect", "/internal/profil")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func DrvUploadPost(c *handler.Context) error {
	err := api_v1.DrvMeldungUpload(c)
	if err != nil {
		return handler.BadRequest("Ein Fehler ist aufgetreten! Bitte versuche es erneut.")
	}

	return handler.OK("Upload erfolgreich!")
}

func SetzungsVerwaltungLosungPost(c *handler.Context) error {
	err := api_v1.SetzungsLosung(c)
	if err != nil {
		return handler.BadRequest(fmt.Sprintf("Ein Fehler ist aufgetreten: %s", err.Error()))
	}
	return handler.OK("Losung erfolgreich!")
}

func SetzungsVerwaltungLosungDelete(c *handler.Context) error {
	err := api_v1.ResetSetzung(c)
	if err != nil {
		return handler.BadRequest(fmt.Sprintf("Ein Fehler ist aufgetreten: %s", err.Error()))
	}
	return handler.OK("Setzung erfolgreich zurückgesetzt!")
}

func SetzungsVerwaltungAenderungRennenPost(c *handler.Context) error {
	var (
		err    error
		rUuid  uuid.UUID
		rennen crud.Rennen
	)

	rUuid, err = uuid.Parse(c.Param("param"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	rennen, err = crud.GetRennen(c.Request.Context(), rUuid)
	if err != nil {
		return handler.NotFound("Rennen nicht gefunden")
	}

	payloadStr := c.FormValue("params")
	payload := make(map[string]any)
	err = json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		return handler.NotAcceptable("Invalid JSON")
	}

	meldOrderLs, ok := payload["order"].([]any)
	if !ok {
		return handler.BadRequest("Order nicht gefunden")
	}
	abteilungParam := payload["abteilung"]
	if abteilungParam == nil {
		return handler.NotAcceptable("Abteilung nicht gefunden")
	}
	targetAbteilung := int32(abteilungParam.(float64))

	for i, m := range meldOrderLs {
		mUuid, err := uuid.Parse(m.(string))
		if err != nil {
			return handler.NotAcceptable("Invalid UUID")
		}

		for _, meldung := range rennen.Meldungen {
			if meldung.Uuid == mUuid {
				bahn := int32(i) + 1

				err = crud.UpdateMeldungSetzung(c.Request.Context(), sqlc.UpdateMeldungSetzungParams{
					Uuid:      meldung.Uuid,
					Abteilung: targetAbteilung,
					Bahn:      bahn,
				})
				if err != nil {
					return handler.InternalError("Error while updating meldung setzung")
				}
				continue
			}
		}
	}

	return handler.OK("Setzung erfolgreich!")
}

func PausenNew(c *handler.Context) error {
	nachRennenUuidStr := c.Param("nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	p := crud.Pause{Pause: sqlc.Pause{ID: int32(-2), NachRennenUuid: nachRennenUuid, Laenge: 45}}

	templ.Handler(ui_pages.PausenPausenEntry(p)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func PausenPost(c *handler.Context) error {
	idStr := c.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		return handler.NotAcceptable("Invalid ID")
	}
	laengeStr := c.FormValue("laenge")
	laenge, err := strconv.Atoi(laengeStr)
	if err != nil || laenge < 0 || laenge > 120 {
		slog.Error(fmt.Sprintf("Laenge: %s - Error: %s", laengeStr, err.Error()))
		return handler.NotAcceptable("Invalid laenge")
	}
	nachRennenUuidStr := c.FormValue("nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		slog.Error(fmt.Sprintf("UUID: %s - Error: %s", nachRennenUuidStr, err.Error()))
		return handler.NotAcceptable("Invalid UUID")
	}

	if id == -2 {
		_, err = crud.CreatePause(c.Request.Context(), sqlc.CreatePauseParams{
			NachRennenUuid: nachRennenUuid,
			Laenge:         int32(laenge),
		})
		if err != nil {
			return handler.InternalError("Error while creating pause")
		}

		templ.Handler(ui_pages.InternalRegattaleitungPausen()).ServeHTTP(c.Writer, c.Request)
		return nil
	} else {
		_, err = crud.UpdatePause(c.Request.Context(), sqlc.UpdatePauseParams{
			ID:     int32(id),
			Laenge: int32(laenge),
		})
		if err != nil {
			return handler.InternalError("Error while updating pause")
		}

		templ.Handler(ui_pages.InternalRegattaleitungPausen()).ServeHTTP(c.Writer, c.Request)
		return nil
	}
}

func PausenDelete(c *handler.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		return handler.NotAcceptable("Invalid ID")
	}

	err = crud.DeletePause(c.Request.Context(), int32(id))
	if err != nil {
		return handler.InternalError("Error while deleting pause")
	}

	templ.Handler(ui_pages.InternalRegattaleitungPausen()).ServeHTTP(c.Writer, c.Request)
	return nil
}

func ZeitplanPost(c *handler.Context) error {
	startzeit_saStr := c.FormValue("startzeit_sa")
	startzeit_soStr := c.FormValue("startzeit_so")

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
		return handler.BadRequest("Ungültige Startzeit").WithForm(ui_pages.InternalRegattaleitungZeitplan("", fieldErrors))
	}

	zeitplan := service.SetZeitplanParams{
		SaStartStunde: startzeit_sa,
		SoStartStunde: startzeit_so,
	}

	err = service.SetZeitplan(c.Request.Context(), zeitplan)
	if err != nil {
		return handler.InternalError("Error while creating zeitplan")
	}

	return handler.OK("Zeitplan erstellt")
}

func StartnummernVerteilenPost(c *handler.Context) error {
	err := service.SetStartnummern(c.Request.Context())
	if err != nil {
		return handler.InternalError(fmt.Sprintf("Error while setting startnummern: %s", err.Error()))
	}

	return handler.OK("Startnummern erfolgreich verteilt!")
}

func StartnummernVerteilenDelete(c *handler.Context) error {
	err := service.ResetStartnummern(c.Request.Context())
	if err != nil {
		return handler.InternalError("Error while resetting startnummern")
	}

	return handler.OK("Startnummern erfolgreich zurückgesetzt!")
}

func PdfMeldeergebnisPost(c *handler.Context) error {
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
		return handler.InternalError(fmt.Sprintf("Fehler während PDF Erstellung: %s", err.Error()))
	}

	templ.Handler(ui_pages.InternalRegattaleitungPdfMeldeergebnis(true)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func AbmeldungDelete(c *handler.Context) error {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	meldungUuidStr := c.Param("m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading verein")
	}
	meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return handler.InternalError("Error while loading meldung")
	}

	if meldung.VereinUuid != verein.Uuid {
		return handler.NotAcceptable("Invalid UUID")
	}

	err = crud.Abmeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return handler.InternalError("Error while deleting meldung")
	}

	c.Writer.Header().Set("HX-Redirect", fmt.Sprintf("/internal/regattabuero/%s/abmeldung", verein.Uuid))
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func UmmeldungPost(c *handler.Context) error {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	meldungUuidStr := c.Param("m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading verein")
	}
	meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return handler.InternalError("Error while loading meldung")
	}

	if meldung.VereinUuid != verein.Uuid {
		return handler.NotAcceptable("Invalid UUID")
	}

	athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading athleten")
	}

	if err := c.Request.ParseForm(); err != nil {
		return handler.BadRequest("Error parsing form")
	}

	fieldErrors := make(map[string]string)
	for i := range meldung.Athleten {
		athUuidStr := c.Request.FormValue(fmt.Sprintf("athleten_%d", i))
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
		err = crud.Ummeldung(c.Request.Context(), sqlc.UmmeldungParams{
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
		return handler.BadRequest("Fehler bei der Ummeldung").WithForm(ui_pages.InternalRegattabueroUmmeldungMeldung(verein, meldung, athleten, "", fieldErrors))
	}

	meldungen, err := crud.GetAllMeldungForVerein(c.Request.Context(), verein.Uuid)
	if err != nil {
		return handler.InternalError("Error while loading meldungen")
	}

	c.Writer.Header().Set("HX-Push-Url", fmt.Sprintf("/internal/regattabuero/%s/ummeldung", verein.Uuid))
	c.Writer.WriteHeader(http.StatusOK)
	return ui_pages.InternalRegattabueroUmmeldung(verein, meldungen).Render(context.Background(), c.Writer)
}

func NachmeldungPost(c *handler.Context) error {
	vereinUuidStr := c.Param("v_uuid")
	rennenUuidStr := c.Param("r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	rennen, err := crud.GetRennen(c.Request.Context(), rennenUuid)
	if err != nil {
		return handler.InternalError("Error while loading rennen")
	}

	if err := c.Request.ParseForm(); err != nil {
		return handler.BadRequest("Error parsing form")
	}

	vereinUuid, err := uuid.Parse(c.Request.FormValue("verein_uuid"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading verein")
	}

	athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading athleten")
	}

	numAthletes, stmRequired := rennen.GetTeilnehmerMeldeParams()

	params := api_v1.PostNachmeldungParams{
		VereinUuid:                    c.Request.FormValue("verein_uuid"),
		RennenUuid:                    c.Request.FormValue("rennen_uuid"),
		DoppeltesMeldentgeldBefreiung: c.Request.FormValue("doppeltes_meldentgeld_befreiung") != "",
		Athleten:                      []api_v1.PostNachmeldungAthletParams{},
	}

	fieldErrors := make(map[string]string)
	hasAthlete := false
	for i := range numAthletes {
		athVal := c.Request.FormValue(fmt.Sprintf("athleten_%d", i))
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
		stmVal := c.Request.FormValue("stm")
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
		return handler.BadRequest("Bitte wähle mindestens einen Teilnehmer aus").WithForm(ui_pages.InternalRegattabueroNachmeldungMeldung(verein, rennen, athleten, "", fieldErrors))
	}

	m, err := api_v1.CreateNachmeldung(c.Request.Context(), params)
	if err != nil {
		return handler.InternalError("Error creating nachmeldung: " + err.Error())
	}
	meldung, err := crud.GetMeldung(c.Request.Context(), m.Uuid)
	if err != nil {
		return handler.InternalError("Error while loading meldung")
	}

	c.Writer.Header().Set("HX-Push-Url", fmt.Sprintf("/internal/regattabuero/%s/nachmeldung/success/%s", vereinUuidStr, m.Uuid.String()))
	c.Writer.WriteHeader(http.StatusOK)
	return ui_pages.InternalRegattabueroNachmeldungSuccess(meldung).Render(context.Background(), c.Writer)
}

func RennenTab(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	showEmpty := c.GetQueryParam("show_empty") == "true"
	showStarted := c.GetQueryParam("show_started") == "true"
	urlFormatStr := c.GetQueryParam("url_format_str")

	templ.Handler(ui_components.RennenTab(wettkampf, urlFormatStr, showEmpty, showStarted)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func NewAthletPost(c *handler.Context) error {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while loading verein")
	}

	vorname := c.FormValue("vorname")
	name := c.FormValue("name")
	jahrgang := c.FormValue("jahrgang")
	geschlecht := c.FormValue("geschlecht")
	startberechtigt := c.FormValue("startberechtigt") == "on"

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
		return handler.BadRequest("Bitte alle Pflichtfelder ausfüllen").WithForm(
			ui_pages.InternalRegattabueroNewAthlet(verein, "", fieldErrors),
		)
	}

	athletUuid, err := uuid.NewV7()
	if err != nil {
		return handler.InternalError("Error generating UUID")
	}

	a, err := crud.CreateAthlet(c.Request.Context(), sqlc.CreateAthletParams{
		Uuid:            athletUuid,
		VereinUuid:      vereinUuid,
		Name:            name,
		Vorname:         vorname,
		Jahrgang:        jahrgang,
		Startberechtigt: startberechtigt,
		Geschlecht:      sqlc.Geschlecht(geschlecht),
	})
	if err != nil {
		return handler.InternalError("Fehler beim Anlegen des Athleten").WithForm(
			ui_pages.InternalRegattabueroNewAthlet(verein, "", nil),
		)
	}

	a.Verein = &verein
	return ui_pages.InternalRegattabueroNewAthletSuccess(a).Render(context.Background(), c.Writer)
}

func WaagePost(c *handler.Context) error {
	err := c.Request.ParseForm()
	if err != nil {
		slog.Error("ParseForm error", "err", err)
		return handler.BadRequest("Fehler beim Verarbeiten der Anfrage")
	}

	idStr := c.Request.FormValue("uuid")
	gewichtStr := c.Request.FormValue("gewicht")

	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Parse UUID error", "err", err)
		return handler.NotAcceptable("Ungültige UUID")
	}

	ath, err := crud.GetAthletMinimal(c.Request.Context(), id)
	if err != nil {
		slog.Error("GetAthletMinimal error", "err", err)
		return err
	}

	fieldErrors := make(map[string]string)
	gewichtFloat, err := strconv.ParseFloat(gewichtStr, 32)
	if err != nil {
		fieldErrors["gewicht"] = "Ungültiges Gewicht"
	}
	gewicht := int(gewichtFloat * 10)

	if len(fieldErrors) > 0 {
		return handler.BadRequest("Ungültiges Gewicht").WithForm(ui_pages.InternalRegattabueroWaage(ath, "", fieldErrors))
	}

	err = ath.UpdateGewicht(c.Request.Context(), gewicht)
	if err != nil {
		slog.Error("UpdateGewicht error", "err", err)
		return handler.InternalError("Fehler beim Aktualisieren des Gewichts")
	}

	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.NotFound("Verein nicht gefunden")
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(c.Request.Context(), verein.Uuid)
	if err != nil {
		return handler.InternalError("Error while loading athleten")
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	c.Writer.Header().Set("HX-Push-Url", fmt.Sprintf("/internal/regattabuero/%s/waage", vereinUuidStr))
	c.Writer.WriteHeader(http.StatusOK)
	return ui_pages.InternalRegattabueroWaageWahl(verein, athleten).Render(context.Background(), c.Writer)
}

func ZeitplanCollapseBody(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return handler.NotFound("Wettkampf not found")
	}
	templ.Handler(ui_components.ZeitplanCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func AusschreibungRennenCollapseBody(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return handler.NotFound("Wettkampf not found")
	}
	templ.Handler(ui_pages.AusschreibungRennenCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func MeldeergebnisCollapseBody(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return handler.NotFound("Wettkampf not found")
	}
	templ.Handler(ui_pages.MeldeergebnisCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func MetricsAPIHandler(c *handler.Context) error {
	return api_v1.MetricsApi(c)
}
