package components

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/mailer"
	admin "github.com/bata94/RegattaApi/internal/templates/pages/admin"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func EmailQueueRetry(w http.ResponseWriter, r *http.Request) {
	emailUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	if err := crud.ResetEmailQueue(r.Context(), emailUuid); err != nil {
		webfw.ErrorToast(w, r, "Error while resetting email")
		return
	}

	templ.Handler(admin.EmailQueue()).ServeHTTP(w, r)
}

func EmailQueueDelete(w http.ResponseWriter, r *http.Request) {
	emailUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	if err := crud.DeleteEmailQueue(r.Context(), emailUuid); err != nil {
		webfw.ErrorToast(w, r, "Error while deleting email")
		return
	}

	templ.Handler(admin.EmailQueue()).ServeHTTP(w, r)
}

const attachmentPrefixLen = 37

func attachmentDisplayName(id string) string {
	if len(id) > attachmentPrefixLen {
		return id[attachmentPrefixLen:]
	}
	return id
}

func saveEmailAttachment(r *http.Request, fh *multipart.FileHeader) (regattaleitung.EmailAttachment, string, error) {
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
	if err := webfw.SaveFile(dest, content); err != nil {
		return regattaleitung.EmailAttachment{}, "", err
	}

	return regattaleitung.EmailAttachment{ID: name, Filename: base}, dest, nil
}

func EmailSendPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if err := r.ParseForm(); err != nil {
			webfw.ErrorToast(w, r, "Fehler beim Verarbeiten des Formulars")
			return
		}
	}

	allObleute := r.FormValue("all_obleute") == "on"
	selectedVereine := r.Form["vereine"]
	manualRecipients := r.FormValue("manual_recipients")
	subject := r.FormValue("subject")
	body := r.FormValue("body")

	var attachments []regattaleitung.EmailAttachment
	var filePaths []string
	if r.MultipartForm != nil {
		for _, fh := range r.MultipartForm.File["files"] {
			att, path, err := saveEmailAttachment(r, fh)
			if err != nil {
				slog.Error("Failed to save uploaded file", "err", err)
				continue
			}
			attachments = append(attachments, att)
			filePaths = append(filePaths, path)
		}
	}

	for _, id := range r.Form["saved_files"] {
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
		obleute, err := crud.GetAllObmann(r.Context())
		if err != nil {
			webfw.ErrorToast(w, r, "Fehler beim Laden der Obleute")
			return
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
		obleute, err := crud.GetAllObmannForVerein(r.Context(), vUuid)
		if err != nil {
			webfw.ErrorToast(w, r, "Fehler beim Laden der Obleute")
			return
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
		webfw.ErrorWithForm(w, r, regattaleitung.EmailCompose(subject, body, manualRecipients, selectedVereine, allObleute, fieldErrors, attachments), "Validation error")
		return
	}

	to := make([]string, 0, len(recipients))
	for addr := range recipients {
		to = append(to, addr)
	}

	if err := mailer.Enqueue(r.Context(), mailer.Params{
		To:      to,
		Subject: subject,
		Body:    body,
		Files:   filePaths,
	}); err != nil {
		slog.Error("Email enqueue failed", "err", err)
		webfw.ErrorToast(w, r, "Fehler beim Einreihen der E-Mail")
		return
	}

	templ.Handler(regattaleitung.EmailCompose("", "", "", nil, false, nil, nil)).ServeHTTP(w, r)

	oobToast := fmt.Sprintf(
		`<div hx-swap-oob="beforeend:#toast-container"><div class="alert alert-success flex flex-row justify-between items-center gap-2"><span>%s</span><button class="btn btn-sm btn-circle btn-ghost" onclick="this.parentElement.remove()">✕</button></div></div>`,
		"E-Mail wurde in die Warteschlange gelegt",
	)
	if _, err := fmt.Fprint(w, oobToast); err != nil {
		slog.Error("Error writing OOB toast", "err", err)
	}
}
