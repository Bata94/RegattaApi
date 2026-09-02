package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func TestSmokeAllGetRoutes(t *testing.T) {
	_ = godotenv.Load()
	config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.C.DB.Host, config.C.DB.Port, config.C.DB.User, config.C.DB.Name, config.C.DB.Password, config.C.DB.SSLMode)

	probe, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("smoke: dev DB not reachable: %v", err)
	}
	_ = probe.Close(context.Background())

	DB.InitConnection(DB.DBServerOptions{
		Host:     config.C.DB.Host,
		Port:     config.C.DB.Port,
		User:     config.C.DB.User,
		Password: config.C.DB.Password,
		Name:     config.C.DB.Name,
		Sslmode:  config.C.DB.SSLMode,
	})
	defer func() {
		if err := DB.ShutdownConnection(); err != nil {
			t.Logf("smoke: db shutdown error: %v", err)
		}
	}()

	router := newChiRouter()
	srv := httptest.NewServer(router)
	defer srv.Close()

	token := smokeLogin(srv.URL)
	if token == "" {
		t.Skip("smoke: could not login as admin/admin")
	}

	var patterns []string
	if err := chi.Walk(router, func(method, pattern string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if method == http.MethodGet {
			patterns = append(patterns, pattern)
		}
		return nil
	}); err != nil {
		t.Fatalf("smoke: walk router: %v", err)
	}

	res := &paramsResolver{
		ctx:   context.Background(),
		cache: map[entityKind]string{},
	}
	client := &http.Client{
		Timeout:       45 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	tested, skipped := 0, 0
	for _, pattern := range patterns {
		if skipSmokeRoute(pattern) {
			skipped++
			continue
		}

		path, ok := fillRoute(pattern, res)
		if !ok {
			skipped++
			t.Logf("skip  %-58s (no data)", pattern)
			continue
		}
		tested++

		req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
		if err != nil {
			t.Errorf("smoke: GET %s: %v", pattern, err)
			continue
		}
		req.Header.Set("Cookie", "auth_token="+token)
		if strings.HasPrefix(pattern, "/comp/") {
			req.Header.Set("HX-Request", "true")
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("smoke: GET %s (-> %s): %v", pattern, path, err)
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode >= 500 {
			t.Errorf("smoke: GET %s (-> %s): status %d", pattern, path, resp.StatusCode)
			continue
		}
		t.Logf("ok    %-58s -> %3d %s", pattern, resp.StatusCode, path)
	}

	t.Logf("smoke: %d GET routes tested, %d skipped", tested, skipped)
}

func smokeLogin(base string) string {
	resp, err := http.PostForm(base+"/login", url.Values{
		"username": {"admin"},
		"password": {"admin"},
	})
	if err != nil {
		return ""
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	for _, c := range resp.Cookies() {
		if c.Name == "auth_token" && c.Value != "" {
			return c.Value
		}
	}
	return ""
}

func skipSmokeRoute(pattern string) bool {
	if strings.Contains(pattern, "*") {
		return true
	}
	switch pattern {
	case "/ws/zeitnahme", "/logout":
		return true
	}
	return false
}

type entityKind int

const (
	kindUnknown entityKind = iota
	kindVerein
	kindAthlet
	kindMeldung
	kindRennen
	kindUser
	kindUsersGroup
	kindObmann
	kindPauseID
	kindUsername
	kindUsersGroupName
	kindWettkampf
	kindMeldeergebnisFile
)

func entityKindFor(pattern, param string) entityKind {
	switch param {
	case "v_uuid":
		return kindVerein
	case "a_uuid":
		return kindAthlet
	case "m_uuid":
		return kindMeldung
	case "r_uuid", "nach_rennen_uuid", "param":
		return kindRennen
	case "wettkampf":
		return kindWettkampf
	case "id":
		return kindPauseID
	case "filename":
		return kindMeldeergebnisFile
	case "name":
		if strings.Contains(pattern, "users/group") {
			return kindUsersGroupName
		}
		return kindUsername
	case "uuid":
		switch {
		case strings.Contains(pattern, "profil/password"):
			return kindUser
		case strings.Contains(pattern, "meldung/verein"):
			return kindVerein
		case strings.Contains(pattern, "users/group"):
			return kindUsersGroup
		case strings.Contains(pattern, "usergroups"):
			return kindUsersGroup
		case strings.Contains(pattern, "vereine"):
			return kindVerein
		case strings.Contains(pattern, "obleute"):
			return kindObmann
		case strings.Contains(pattern, "/admin/user"):
			return kindUser
		case strings.Contains(pattern, "rechnung"):
			return kindVerein
		case strings.Contains(pattern, "meldung"):
			return kindMeldung
		case strings.Contains(pattern, "rennen"):
			return kindRennen
		case strings.Contains(pattern, "verein"):
			return kindVerein
		case strings.Contains(pattern, "athlet"):
			return kindAthlet
		case strings.Contains(pattern, "users"):
			return kindUser
		}
	}
	return kindUnknown
}

func fillRoute(pattern string, res *paramsResolver) (string, bool) {
	parts := strings.Split(pattern, "/")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			name := p[1 : len(p)-1]
			v, ok := res.resolve(entityKindFor(pattern, name))
			if !ok {
				return "", false
			}
			out = append(out, v)
			continue
		}
		if strings.Contains(p, "*") {
			return "", false
		}
		out = append(out, p)
	}
	return strings.Join(out, "/"), true
}

type paramsResolver struct {
	ctx   context.Context
	cache map[entityKind]string
}

func (res *paramsResolver) resolve(kind entityKind) (string, bool) {
	if v, ok := res.cache[kind]; ok {
		return v, true
	}
	v, ok := res.fetch(kind)
	if ok {
		res.cache[kind] = v
	}
	return v, ok
}

func (res *paramsResolver) fetch(kind entityKind) (string, bool) {
	ctx := res.ctx

	switch kind {
	case kindVerein:
		vs, err := crud.GetAllVerein(ctx)
		if err != nil || len(vs) == 0 {
			return "", false
		}
		return vs[0].Uuid.String(), true
	case kindAthlet:
		as, err := crud.GetAllAthlet(ctx)
		if err != nil || len(as) == 0 {
			return "", false
		}
		return as[0].Uuid.String(), true
	case kindMeldung:
		ms, err := crud.GetAllMeldungen(ctx)
		if err != nil || len(ms) == 0 {
			return "", false
		}
		return ms[0].Uuid.String(), true
	case kindRennen:
		rs, err := crud.GetAllRennen(ctx, crud.GetAllRennenParams{
			ShowEmpty:   true,
			ShowStarted: true,
		})
		if err != nil || len(rs) == 0 {
			return "", false
		}
		return rs[0].Uuid.String(), true
	case kindUser:
		us, err := crud.GetAllUsers(ctx)
		if err != nil || len(us) == 0 {
			return "", false
		}
		return us[0].Uuid.String(), true
	case kindUsersGroup:
		gs, err := crud.GetAllUsersGroups(ctx)
		if err != nil || len(gs) == 0 {
			return "", false
		}
		return gs[0].Uuid.String(), true
	case kindObmann:
		os, err := crud.GetAllObmann(ctx)
		if err != nil || len(os) == 0 {
			return "", false
		}
		return os[0].Uuid.String(), true
	case kindPauseID:
		ps, err := crud.GetAllPausen(ctx)
		if err != nil || len(ps) == 0 {
			return "", false
		}
		return strconv.Itoa(int(ps[0].ID)), true
	case kindUsername:
		us, err := crud.GetAllUsers(ctx)
		if err != nil || len(us) == 0 {
			return "", false
		}
		return us[0].Username, true
	case kindUsersGroupName:
		gs, err := crud.GetAllUsersGroups(ctx)
		if err != nil || len(gs) == 0 {
			return "", false
		}
		return gs[0].Name, true
	case kindWettkampf:
		if len(crud.AllWettkampf) == 0 {
			return "", false
		}
		return string(crud.AllWettkampf[0]), true
	case kindMeldeergebnisFile:
		files, err := filepath.Glob(filepath.Join(config.C.Paths.FilesDir, "meldeergebnis", "*"))
		if err != nil || len(files) == 0 {
			return "", false
		}
		return filepath.Base(files[0]), true
	}

	return "", false
}
