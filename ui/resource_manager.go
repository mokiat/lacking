package ui

import (
	"fmt"
	"image"
	"io"

	_ "image/jpeg"
	_ "image/png"

	"github.com/mokiat/lacking/util/resource"
	"golang.org/x/image/font/opentype"
)

func newResourceManager(locator resource.ReadLocator, imgFact *imageFactory, fntFact *fontFactory) *resourceManager {
	return &resourceManager{
		locator: locator,
		imgFact: imgFact,
		fntFact: fntFact,
	}
}

type resourceManager struct {
	locator resource.ReadLocator
	imgFact *imageFactory
	fntFact *fontFactory
}

func (m *resourceManager) CreateImage(img image.Image) *Image {
	return m.imgFact.CreateImage(img)
}

func (m *resourceManager) OpenImage(uri string) (*Image, error) {
	fmt.Println("OPENING IMAGE:", uri)
	in, err := m.locator.ReadResource(uri)
	if err != nil {
		return nil, fmt.Errorf("error opening resource: %w", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	return m.CreateImage(img), nil
}

func (m *resourceManager) CreateFont(otFont *opentype.Font) (*Font, error) {
	return m.fntFact.CreateFont(otFont)
}

func (m *resourceManager) OpenFont(uri string) (*Font, error) {
	in, err := m.locator.ReadResource(uri)
	if err != nil {
		return nil, fmt.Errorf("error opening resource: %w", err)
	}
	defer in.Close()

	content, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("error reading resource content: %w", err)
	}

	otFont, err := opentype.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %w", err)
	}
	return m.CreateFont(otFont)
}

func (m *resourceManager) CreateFontCollection(collection *opentype.Collection) (*FontCollection, error) {
	fonts := make([]*Font, collection.NumFonts())
	for i := range fonts {
		otFont, err := collection.Font(i)
		if err != nil {
			return nil, fmt.Errorf("error retrieving font from collection: %w", err)
		}
		font, err := m.CreateFont(otFont)
		if err != nil {
			return nil, fmt.Errorf("error creating font: %w", err)
		}
		fonts[i] = font
	}
	return newFontCollection(fonts), nil
}

func (m *resourceManager) OpenFontCollection(uri string) (*FontCollection, error) {
	in, err := m.locator.ReadResource(uri)
	if err != nil {
		return nil, fmt.Errorf("error opening resource: %w", err)
	}
	defer in.Close()

	content, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("error reading resource content: %w", err)
	}

	otCollection, err := opentype.ParseCollection(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing font collection: %w", err)
	}
	return m.CreateFontCollection(otCollection)
}
