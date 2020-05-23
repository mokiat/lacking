package pack

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Builder interface {
	Build() error
}

func NewPacker() *Packer {
	return &Packer{}
}

type Packer struct {
	resourcesDir string
	assetsDir    string
}

func (p *Packer) SetResourcesDir(resourcesDir string) {
	p.resourcesDir = resourcesDir
}

func (p *Packer) SetAssetsDir(assetsDir string) {
	p.assetsDir = assetsDir
}

func (p *Packer) Store(builders ...Builder) {
	for _, builder := range builders {
		if err := builder.Build(); err != nil {
			log.Fatalf("failed to build asset: %v", err)
		}
	}
}

func (p *Packer) ProgramAssetFile(name string) *ProgramAssetBuilder {
	return &ProgramAssetBuilder{
		Asset: Asset{
			filename: filepath.Join(p.assetsDir, "programs", name),
		},
	}
}

func (p *Packer) TwoDTextureAssetFile(name string) *TwoDTextureAssetBuilder {
	return &TwoDTextureAssetBuilder{
		Asset: Asset{
			filename: filepath.Join(p.assetsDir, "textures", "twod", name),
		},
	}
}

func (p *Packer) ShaderResourceFile(name string) *ShaderResourceFile {
	return &ShaderResourceFile{
		Resource: Resource{
			filename: filepath.Join(p.resourcesDir, "shaders", name),
		},
	}
}

func (p *Packer) ImageResourceFile(name string) *ImageResourceFile {
	return &ImageResourceFile{
		Resource: Resource{
			// TODO: Should be images/twod
			filename: filepath.Join(p.resourcesDir, "textures", "twod", name),
		},
	}
}

type Asset struct {
	filename string
}

func (a Asset) CreateFile() (io.WriteCloser, error) {
	dirname := filepath.Dir(a.filename)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dir %q: %w", dirname, err)
	}

	file, err := os.Create(a.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q: %w", a.filename, err)
	}
	return file, nil
}

type Resource struct {
	filename string
}
