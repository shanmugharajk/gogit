package commit

import (
	"fmt"
	"time"
)

type Author struct {
	Name  string
	Email string
	Time  time.Time
}

func (a *Author) Bytes() []byte {
	return fmt.Appendf(nil, "%s <%s> %d %s",
		a.Name,
		a.Email,
		a.Time.Unix(),
		a.Time.Format("-0700"))
}

func NewAuthor(name string, email string, time time.Time) *Author {
	return &Author{
		Name:  name,
		Email: email,
		Time:  time,
	}
}
