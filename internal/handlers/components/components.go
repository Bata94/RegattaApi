package components

// This file exists to ensure the components package is initialized.
// Handlers have been moved to separate files:
// - auth.go: LoginPost
// - image.go: ImageComponent
// - user.go: UserEditNew, UserEditNewPost
// - password.go: ChangePasswordGet, ChangePasswordPost
// - user_group.go: UserGroupEditNew, UserGroupEditNewPost
// - drv_upload.go: DrvUploadPost
// - setzung.go: SetzungsVerwaltungLosungPost, SetzungsVerwaltungLosungDelete, SetzungsVerwaltungAenderungRennenPost
// - startnummern.go: StartnummernAendernPost, StartnummernBereichPost, StartnummernVerteilenPost, StartnummernVerteilenDelete
// - pausen.go: PausenNew, PausenPost, PausenDelete
// - zeitplan.go: ZeitplanPost
// - pdf.go: PdfMeldeergebnisPost
// - meldeverwaltung.go: AbmeldungDelete, UmmeldungPost, NachmeldungPost
// - athlet.go: NewAthletPost, StartberechtigungPost
// - waage.go: WaagePost
// - collapse.go: RennenTab, ZeitplanCollapseBody, AusschreibungRennenCollapseBody, MeldeergebnisCollapseBody
