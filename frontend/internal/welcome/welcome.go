package welcome

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewWelcomeScreen() fyne.CanvasObject {
	logo := canvas.NewImageFromFile("./Icon.png")
	logo.FillMode = canvas.ImageFillContain
	if fyne.CurrentDevice().IsMobile() {
		logo.SetMinSize(fyne.NewSize(192, 192))
	} else {
		logo.SetMinSize(fyne.NewSize(256, 256))
	}

	authors := widget.NewRichTextFromMarkdown("\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n\nHEEEEELO\n")
	content := container.NewVBox(
		widget.NewLabelWithStyle("Welcome", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		logo,
		container.NewCenter(authors),
	)
	scroll := container.NewScroll(content)

	bgColor := withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
	shadowColor := withAlpha(theme.Color(theme.ColorNameBackground), 0x33)

	underlay := canvas.NewImageFromFile("./Icon.png")
	bg := canvas.NewRectangle(bgColor)
	underlayer := underLayout{}
	slideBG := container.New(underlayer, underlay)
	footerBG := canvas.NewRectangle(shadowColor)

	fyne.CurrentApp().Settings().AddListener(func(fyne.Settings) {
		bgColor = withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
		bg.FillColor = bgColor
		bg.Refresh()

		shadowColor = withAlpha(theme.Color(theme.ColorNameBackground), 0x33)
		footerBG.FillColor = bgColor
	})

	underlay.Resize(fyne.NewSize(1024, 1024))
	scroll.OnScrolled = func(p fyne.Position) {
		underlayer.offset = -p.Y / 3
		underlayer.Layout(slideBG.Objects, slideBG.Size())
	}

	bgClip := container.NewScroll(slideBG)
	bgClip.Direction = container.ScrollNone
	return container.NewStack(container.New(unpad{top: true}, bgClip, bg),
		container.NewBorder(nil,
			container.NewStack(footerBG), nil, nil,
			container.New(unpad{top: true, bottom: true}, scroll)))
}

func withAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: alpha}
}

type underLayout struct {
	offset float32
}

func (u underLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	under := objs[0]
	left := size.Width/2 - under.Size().Width/2
	under.Move(fyne.NewPos(left, u.offset-50))
}

func (u underLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.Size{}
}

type unpad struct {
	top, bottom bool
}

func (u unpad) Layout(objs []fyne.CanvasObject, s fyne.Size) {
	pad := theme.Padding()
	var pos fyne.Position
	if u.top {
		pos = fyne.NewPos(0, -pad)
	}
	size := s
	if u.top {
		size = size.AddWidthHeight(0, pad)
	}
	if u.bottom {
		size = size.AddWidthHeight(0, pad)
	}
	for _, o := range objs {
		o.Move(pos)
		o.Resize(size)
	}
}

func (u unpad) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(100, 100)
}
