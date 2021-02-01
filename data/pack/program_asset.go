package pack

import (
	"fmt"
	"hash"

	"github.com/mokiat/lacking/data/asset"
)

func SaveProgramAsset(uri string, programProvider ProgramProvider) *SaveProgramAssetAction {
	return &SaveProgramAssetAction{
		uri:             uri,
		programProvider: programProvider,
	}
}

var _ Action = (*SaveProgramAssetAction)(nil)

type SaveProgramAssetAction struct {
	uri             string
	programProvider ProgramProvider
}

func (a *SaveProgramAssetAction) Describe() string {
	return fmt.Sprintf("save_program_asset(uri: %q)", a.uri)
}

func (a *SaveProgramAssetAction) Digest(hasher hash.Hash) error {
	return WriteCompositeDigest(hasher, "save_program_asset", HashableParams{
		"uri":     a.uri,
		"program": a.programProvider,
	})
}

func (a *SaveProgramAssetAction) Run(ctx *Context) error {
	logFinished := ctx.LogAction(a.Describe())
	defer logFinished()

	program, err := a.programProvider.Program(ctx)
	if err != nil {
		return fmt.Errorf("failed to get program: %w", err)
	}
	programAsset := &asset.Program{
		VertexSourceCode:   program.VertexShader.Source,
		FragmentSourceCode: program.FragmentShader.Source,
	}

	return ctx.IO(func(storage Storage) error {
		out, err := storage.CreateAsset(a.uri)
		if err != nil {
			return err
		}
		defer out.Close()

		if err := asset.EncodeProgram(out, programAsset); err != nil {
			return fmt.Errorf("failed to encode asset: %w", err)
		}
		return nil
	})
}
