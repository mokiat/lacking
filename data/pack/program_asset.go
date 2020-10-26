package pack

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
)

type SaveProgramAssetAction struct {
	locator         AssetLocator
	uri             string
	programProvider ProgramProvider
}

func (a *SaveProgramAssetAction) Describe() string {
	return fmt.Sprintf("save_program_asset(uri: %q)", a.uri)
}

func (a *SaveProgramAssetAction) Run() error {
	program := a.programProvider.Program()

	programAsset := &asset.Program{
		VertexSourceCode:   program.VertexShader.Source,
		FragmentSourceCode: program.FragmentShader.Source,
	}

	out, err := a.locator.Create(a.uri)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := asset.EncodeProgram(out, programAsset); err != nil {
		return fmt.Errorf("failed to encode asset: %w", err)
	}
	return nil
}
