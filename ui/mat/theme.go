package mat

import "github.com/mokiat/lacking/ui"

var (
	PrimaryColor       ui.Color = ui.RGB(0x21, 0x96, 0xF3)
	PrimaryDarkColor   ui.Color = ui.RGB(0x19, 0x76, 0xD2)
	SecondaryColor     ui.Color = ui.RGB(0x8B, 0xC3, 0x4A)
	SecondaryDarkColor ui.Color = ui.RGB(0x68, 0x9F, 0x38)
	BackgroundColor    ui.Color = ui.RGB(0xFF, 0xFF, 0xFF)
	SurfaceColor       ui.Color = ui.RGB(0xFF, 0xFF, 0xFF)
	ErrorColor         ui.Color = ui.RGB(0xB0, 0x00, 0x20)
	OnPrimaryColor     ui.Color = ui.RGB(0x00, 0x00, 0x00)
	OnSecondaryColor   ui.Color = ui.RGB(0x00, 0x00, 0x00)
	OnSurfaceColor     ui.Color = ui.RGB(0x00, 0x00, 0x00)
	OnErrorColor       ui.Color = ui.RGB(0xFF, 0xFF, 0xFF)

	LightGray ui.Color = ui.RGB(0xEE, 0xEE, 0xEE)
	Gray      ui.Color = ui.RGB(0xDD, 0xDD, 0xDD)
	DarkGray  ui.Color = ui.RGB(0xCC, 0xCC, 0xCC)

	ToolbarHeight      = 64
	ToolbarColor       = SurfaceColor
	ToolbarBorderColor = DarkGray
	ToolbarBorderSize  = 1
	ToolbarItemSpacing = 4
)
