package uuid

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const uuidOID uint32 = 2950

func TestV7Time(t *testing.T) {
	u := NewV7()
	tm := u.Time()
	sec, nsec := tm.UnixTime()
	got := time.Unix(sec, nsec)
	now := time.Now()
	if now.Sub(got) > 2*time.Second || got.Sub(now) > 2*time.Second {
		t.Fatalf("Time() too far off: %v vs %v", got, now)
	}
}

func TestScanValue(t *testing.T) {
	u := NewV7()
	var dst UUID
	if err := dst.Scan(u.String()); err != nil {
		t.Fatal(err)
	}
	if dst != u {
		t.Fatalf("scan mismatch: %v vs %v", dst, u)
	}
	if err := dst.Scan([]byte(u.String())); err != nil {
		t.Fatal(err)
	}
	if dst != u {
		t.Fatalf("byte scan mismatch: %v vs %v", dst, u)
	}
	if err := dst.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if dst != Nil {
		t.Fatal("nil scan should give Nil")
	}
	v, err := u.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != u.String() {
		t.Fatalf("value mismatch: %v", v)
	}
}

func TestParseForms(t *testing.T) {
	u := NewV7()
	for _, s := range []string{u.String(), "{" + u.String() + "}", "urn:uuid:" + u.String(), hexless(u)} {
		p, err := Parse(s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		if p != u {
			t.Fatalf("parse mismatch %q", s)
		}
	}
}

func TestPgxCodecEncodeDecode(t *testing.T) {
	m := pgtype.NewMap()
	u := NewV7()

	var buf []byte
	var err error
	buf, err = m.Encode(uuidOID, pgtype.BinaryFormatCode, u, buf)
	if err != nil {
		t.Fatalf("pgx encode failed: %v", err)
	}

	var dst UUID
	if err := m.Scan(uuidOID, pgtype.BinaryFormatCode, buf, &dst); err != nil {
		t.Fatalf("pgx decode failed: %v", err)
	}
	if dst != u {
		t.Fatalf("roundtrip mismatch: %v vs %v", dst, u)
	}
}

func TestPgxCodecEncodeTextParam(t *testing.T) {
	m := pgtype.NewMap()
	u := NewV7()

	var buf []byte
	var err error
	buf, err = m.Encode(uuidOID, pgtype.TextFormatCode, u, buf)
	if err != nil {
		t.Fatalf("pgx text encode failed: %v", err)
	}
	if string(buf) != u.String() {
		t.Fatalf("text encode mismatch: %q vs %q", string(buf), u.String())
	}
}

func hexless(u UUID) string {
	b, _ := u.MarshalText()
	out := make([]byte, 0, 32)
	for _, c := range b {
		if c != '-' {
			out = append(out, c)
		}
	}
	return string(out)
}
