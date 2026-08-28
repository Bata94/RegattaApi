package uuid

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"

	stduuid "uuid"
)

const g1582ns100 uint64 = 122192928000000000

type UUID stduuid.UUID

var Nil = UUID(stduuid.Nil())

type Time uint64

func New() UUID {
	return UUID(stduuid.New())
}

func NewV7() UUID {
	return UUID(stduuid.NewV7())
}

func Parse(s string) (UUID, error) {
	u, err := stduuid.Parse(s)
	if err != nil {
		return Nil, err
	}
	return UUID(u), nil
}

func MustParse(s string) UUID {
	return UUID(stduuid.MustParse(s))
}

func (u UUID) String() string {
	return stduuid.UUID(u).String()
}

func (u UUID) MarshalText() ([]byte, error) {
	return stduuid.UUID(u).MarshalText()
}

func (u *UUID) UnmarshalText(b []byte) error {
	return (*stduuid.UUID)(u).UnmarshalText(b)
}

func (u UUID) Compare(v UUID) int {
	return stduuid.UUID(u).Compare(stduuid.UUID(v))
}

func (u UUID) Time() Time {
	time := binary.BigEndian.Uint64(u[0:8])
	ver := u[6] >> 4
	switch ver {
	case 6:
		return Time(time)
	case 7:
		return Time((time>>16)*10000 + g1582ns100)
	default:
		t := int64(binary.BigEndian.Uint32(u[0:4]))
		t |= int64(binary.BigEndian.Uint16(u[4:6])) << 32
		t |= int64(binary.BigEndian.Uint16(u[6:8])&0x0fff) << 48
		return Time(t)
	}
}

func (t Time) UnixTime() (sec, nsec int64) {
	sec = int64(t) - int64(g1582ns100)
	nsec = (sec % 10000000) * 100
	sec /= 10000000
	return sec, nsec
}

func (u *UUID) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*u = Nil
	case string:
		p, err := Parse(v)
		if err != nil {
			return err
		}
		*u = p
	case []byte:
		p, err := Parse(string(v))
		if err != nil {
			return err
		}
		*u = p
	default:
		return fmt.Errorf("uuid: cannot scan %T into UUID", src)
	}
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}
