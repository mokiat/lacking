package internal

import "github.com/mokiat/lacking/ui"

var _ ui.Font = (*Font)(nil)

type Font struct{}

func (f *Font) Family() string {
	return ""
}

func (f *Font) SubFamily() string {
	return ""
}
