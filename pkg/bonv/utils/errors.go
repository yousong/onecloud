package utils

import (
	"bytes"
)

type ErrList []error

func (el ErrList) Error() string {
	b := bytes.Buffer{}
	for _, err := range el {
		b.WriteString(err.Error())
		b.WriteRune('\n')
	}
	return b.String()
}

func (el ErrList) ToError() error {
	if len(el) == 0 {
		return nil
	}
	return el
}
