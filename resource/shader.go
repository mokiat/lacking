package resource

import (
	"crypto/sha256"
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/graphics"
)

type ShaderInfo struct {
	Type                string
	HasMetalnessTexture bool
	HasRoughnessTexture bool
	HasAlbedoTexture    bool
	HasNormalTexture    bool
}

func (i ShaderInfo) ID() string {
	digest := sha256.New()
	fmt.Fprintf(digest, "%d%s", len(i.Type), i.Type)
	fmt.Fprintf(digest, "1%t", i.HasMetalnessTexture)
	fmt.Fprintf(digest, "1%t", i.HasRoughnessTexture)
	fmt.Fprintf(digest, "1%t", i.HasAlbedoTexture)
	fmt.Fprintf(digest, "1%t", i.HasNormalTexture)
	return fmt.Sprintf("%x", digest.Sum(nil))
}

type Shader struct {
	Info            ShaderInfo
	GeometryProgram *graphics.Program
	ForwardProgram  *graphics.Program
}

func NewShaderOperator(gfxWorker *async.Worker) *ShaderOperator {
	return &ShaderOperator{
		gfxWorker: gfxWorker,
	}
}

type ShaderOperator struct {
	gfxWorker *async.Worker
}

func (o *ShaderOperator) Allocator(info ShaderInfo) Allocator {
	switch info.Type {
	case "pbr":
		return o.pbrAllocator(info)
	default:
		return AllocatorFunc(func(set *Set) (interface{}, error) {
			return nil, fmt.Errorf("unsupported shader type: %s", info.Type)
		})
	}
}

func (o *ShaderOperator) pbrAllocator(info ShaderInfo) Allocator {
	return AllocatorFunc(func(set *Set) (interface{}, error) {
		shader := &Shader{
			Info:            info,
			GeometryProgram: &graphics.Program{},
		}

		gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
			spec := PBRGeometrySpec{
				UsesAlbedoTexture: info.HasAlbedoTexture,
				UsesTexCoord0:     info.HasAlbedoTexture,
			}
			return shader.GeometryProgram.Allocate(graphics.ProgramData{
				VertexShaderSourceCode:   BuildPBRGeometryVertexShader(spec),
				FragmentShaderSourceCode: BuildPBRGeometryFragmentShader(spec),
			})
		}))
		if err := gfxTask.Wait().Err; err != nil {
			return nil, fmt.Errorf("failed to allocate gfx program: %w", err)
		}
		return shader, nil
	})
}

func (o *ShaderOperator) Releaser() Releaser {
	return ReleaserFunc(func(resource interface{}) error {
		shader := resource.(*Shader)

		if shader.GeometryProgram != nil {
			gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
				return shader.GeometryProgram.Release()
			}))
			if err := gfxTask.Wait().Err; err != nil {
				return fmt.Errorf("failed to release gfx program: %w", err)
			}
		}
		if shader.ForwardProgram != nil {
			gfxTask := o.gfxWorker.Schedule(async.VoidTask(func() error {
				return shader.ForwardProgram.Release()
			}))
			if err := gfxTask.Wait().Err; err != nil {
				return fmt.Errorf("failed to release gfx program: %w", err)
			}
		}
		return nil
	})
}
