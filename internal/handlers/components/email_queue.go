package components

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/mailer"
	admin "github.com/bata94/RegattaApi/internal/templates/pages/admin"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/google/uuid"
)

func EmailQueueRetry(c *handler.Context) error {
	uuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	if err := crud.ResetEmailQueue(c.Request.Context(), uuid); err != nil {
		return handler.InternalError("Error while resetting email")
	}

	templ.Handler(admin.EmailQueue()).ServeHTTP(c.Writer, c.Request)
	return nil
}

func EmailQueueDelete(c *handler.Context) error {
	uuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	if err := crud.DeleteEmailQueue(c.Request.Context(), uuid); err != nil {
		return handler.InternalError("Error while deleting email")
	}

	templ.Handler(admin.EmailQueue()).ServeHTTP(c.Writer, c.Request)
	return nil
}

// attachmentPrefixLen is the length of the "<uuidv7>_" prefix prepended to
// every persisted attachment filename (36 char uuid string + underscore).
const attachmentPrefixLen = 37

func attachmentDisplayName(id string) string {
	if len(id) > attachmentPrefixLen {
		return id[attachmentPrefixLen:]
	}
	return id
}

func saveEmailAttachment(c *handler.Context, fh *multipart.FileHeader) (regattaleitung.EmailAttachment, string, error) {
	f, err := fh.Open()
	if err != nil {
		return regattaleitung.EmailAttachment{}, "", err
	}
	content, err := io.ReadAll(f)
	if closeErr := f.Close(); closeErr != nil {
		slog.Error("Failed to close uploaded file", "err", closeErr)
	}
	if err != nil {
		return regattaleitung.EmailAttachment{}, "", err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return regattaleitung.EmailAttachment{}, "", err
	}

	base := filepath.Base(fh.Filename)
	name := id.String() + "_" + base
	dest := filepath.Join(config.C.Paths.FilesDir, "email_attachments", name)
	if err := c.SaveFile(dest, content); err != nil {
		return regattaleitung.EmailAttachment{}, "", err
	}

	return regattaleitung.EmailAttachment{ID: name, Filename: base}, dest, nil
}

func EmailSendPost(c *handler.Context) error {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		if err := c.Request.ParseForm(); err != nil {
			return handler.BadRequest("Fehler beim Verarbeiten des Formulars")
		}
	}

	allObleute := c.FormValue("all_obleute") == "on"
	selectedVereine := c.Request.Form["vereine"]
	manualRecipients := c.FormValue("manual_recipients")
	subject := c.FormValue("subject")
	body := c.FormValue("body")

	// Persist newly uploaded files immediately so they survive a validation error.
	var attachments []regattaleitung.EmailAttachment
	var filePaths []string
	if c.Request.MultipartForm != nil {
		for _, fh := range c.Request.MultipartForm.File["files"] {
			att, path, err := saveEmailAttachment(c, fh)
			if err != nil {
				slog.Error("Failed to save uploaded file", "err", err)
				continue
			}
			attachments = append(attachments, att)
			filePaths = append(filePaths, path)
		}
	}

	// Previously persisted attachments (hidden inputs) survive re-submits.
	for _, id := range c.Request.Form["saved_files"] {
		id = filepath.Base(strings.TrimSpace(id))
		if id == "" || id == "." {
			continue
		}
		filePaths = append(filePaths, filepath.Join(config.C.Paths.FilesDir, "email_attachments", id))
		attachments = append(attachments, regattaleitung.EmailAttachment{ID: id, Filename: attachmentDisplayName(id)})
	}

	fieldErrors := make(map[string]string)
	if subject == "" {
		fieldErrors["subject"] = "Betreff erforderlich"
	}
	if body == "" {
		fieldErrors["body"] = "Nachricht erforderlich"
	}

	recipients := map[string]struct{}{}

	if allObleute {
		obleute, err := crud.GetAllObmann(c.Request.Context())
		if err != nil {
			return handler.InternalError("Fehler beim Laden der Obleute")
		}
		for _, o := range obleute {
			if o.Email.Valid {
				recipients[o.Email.String] = struct{}{}
			}
		}
	}

	for _, vStr := range selectedVereine {
		vUuid, err := uuid.Parse(vStr)
		if err != nil {
			continue
		}
		obleute, err := crud.GetAllObmannForVerein(c.Request.Context(), vUuid)
		if err != nil {
			return handler.InternalError("Fehler beim Laden der Obleute")
		}
		for _, o := range obleute {
			if o.Email.Valid {
				recipients[o.Email.String] = struct{}{}
			}
		}
	}

	for _, addr := range strings.FieldsFunc(manualRecipients, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	}) {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			recipients[addr] = struct{}{}
		}
	}

	if len(recipients) == 0 {
		fieldErrors["manual_recipients"] = "Mindestens ein Empfänger erforderlich"
	}

	if len(fieldErrors) > 0 {
		return handler.ValidationError(fieldErrors).WithForm(regattaleitung.EmailCompose(subject, body, manualRecipients, selectedVereine, allObleute, fieldErrors, attachments))
	}

	to := make([]string, 0, len(recipients))
	for addr := range recipients {
		to = append(to, addr)
	}

	if err := mailer.Enqueue(c.Request.Context(), mailer.Params{
		To:      to,
		Subject: subject,
		Body:    body,
		Files:   filePaths,
	}); err != nil {
		slog.Error("Email enqueue failed", "err", err)
		return handler.InternalError("Fehler beim Einreihen der E-Mail")
	}

	templ.Handler(regattaleitung.EmailCompose("", "", "", nil, false, nil, nil)).ServeHTTP(c.Writer, c.Request)

	oobToast := fmt.Sprintf(
		`<div hx-swap-oob="beforeend:#toast-container"><div class="alert alert-success flex flex-row justify-between items-center gap-2"><span>%s</span><button class="btn btn-sm btn-circle btn-ghost" onclick="this.parentElement.remove()">✕</button></div></div>`,
		"E-Mail wurde in die Warteschlange gelegt",
	)
	if _, err := fmt.Fprint(c.Writer, oobToast); err != nil {
		slog.Error("Error writing OOB toast", "err", err)
	}

	return nil
}
