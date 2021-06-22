package internal

import (
	"strings"

	"github.com/mokiat/lacking/ui"
)

func NewFont(familyName, subFamilyName string) *Font {
	return &Font{
		familyName:    strings.ToLower(familyName),
		subFamilyName: strings.ToLower(subFamilyName),
	}
}

var _ ui.Font = (*Font)(nil)

type Font struct {
	familyName    string
	subFamilyName string
}

func (f *Font) Family() string {
	return f.familyName
}

func (f *Font) SubFamily() string {
	return f.subFamilyName
}
