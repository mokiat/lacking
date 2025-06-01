package game

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/render"
	"golang.org/x/sync/errgroup"
)

type ArmatureTemplate struct {
	Definition   *graphics.ArmatureDefinition
	NodeBindings []uint32
}

func (l *AssetLoader) ResolveArmatureTemplate(assetArmature dto.Armature) (Identifiable[ArmatureTemplate], error) {
	info := graphics.ArmatureDefinitionInfo{
		InverseBindMatrices: make([]sprec.Mat4, len(assetArmature.Joints)),
	}
	for i, assetJoint := range assetArmature.Joints {
		info.InverseBindMatrices[i] = assetJoint.InverseBindMatrix
	}
	var armatureDefinition *graphics.ArmatureDefinition
	allocateDefinition := func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		armatureDefinition = gfxEngine.CreateArmatureDefinition(info)
		return nil
	}
	if err := l.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[ArmatureTemplate]{}, err
	}

	nodeBindings := make([]uint32, len(assetArmature.Joints))
	for i, assetJoint := range assetArmature.Joints {
		nodeBindings[i] = assetJoint.NodeID
	}

	return Identifiable[ArmatureTemplate]{
		ID: assetArmature.ID,
		Value: ArmatureTemplate{
			Definition:   armatureDefinition,
			NodeBindings: nodeBindings,
		},
	}, nil
}

func (l *AssetLoader) ResolveArmatureTemplates(assetArmatures []dto.Armature) (IdentifiableList[ArmatureTemplate], error) {
	templates := make(IdentifiableList[ArmatureTemplate], len(assetArmatures))
	var group errgroup.Group
	for i, assetArmature := range assetArmatures {
		group.Go(func() error {
			template, err := l.ResolveArmatureTemplate(assetArmature)
			templates[i] = template
			return err
		})
	}
	return templates, group.Wait()
}

func (s *Scene) InstantiateArmatureTemplate(template ArmatureTemplate, nodes IdentifiableList[*hierarchy.Node]) *graphics.Armature {
	armature := s.gfxScene.CreateArmature(graphics.ArmatureInfo{
		Definition: template.Definition,
	})
	for j, nodeID := range template.NodeBindings {
		if jointNode, ok := nodes.FindByID(nodeID); ok {
			jointNode.SetTarget(BoneNodeTarget{
				Armature:  armature,
				BoneIndex: j,
			})
		}
	}
	return armature
}

func (l *AssetLoader) ResolveMeshGeomety(assetGeometry dto.Geometry) (Identifiable[*graphics.MeshGeometry], error) {
	meshFragmentsInfo := make([]graphics.MeshGeometryFragmentInfo, len(assetGeometry.Fragments))
	for j, assetFragment := range assetGeometry.Fragments {
		meshFragmentsInfo[j] = graphics.MeshGeometryFragmentInfo{
			Name:            assetFragment.Name,
			Topology:        l.resolveTopology(assetFragment.Topology),
			IndexByteOffset: assetFragment.IndexByteOffset,
			IndexCount:      assetFragment.IndexCount,
		}
	}

	meshGeometryInfo := graphics.MeshGeometryInfo{
		VertexBuffers: gog.Map(assetGeometry.VertexBuffers, func(buffer dto.VertexBuffer) graphics.MeshGeometryVertexBuffer {
			return graphics.MeshGeometryVertexBuffer{
				ByteStride: buffer.Stride,
				Data:       buffer.Data,
			}
		}),
		VertexFormat: l.resolveVertexFormat(assetGeometry.VertexLayout),
		IndexBuffer: graphics.MeshGeometryIndexBuffer{
			Data:   assetGeometry.IndexBuffer.Data,
			Format: l.resolveIndexFormat(assetGeometry.IndexBuffer.IndexLayout),
		},
		Fragments:            meshFragmentsInfo,
		BoundingSphereRadius: assetGeometry.BoundingSphereRadius,
		MinDistance:          opt.V(assetGeometry.MinDistance),
		MaxDistance:          opt.V(assetGeometry.MaxDistance),
		MaxCascade:           opt.V(assetGeometry.MaxCascade),
	}

	var meshGeometry *graphics.MeshGeometry
	allocateGeometry := func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		meshGeometry = gfxEngine.CreateMeshGeometry(meshGeometryInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateGeometry).Wait(); err != nil {
		return Identifiable[*graphics.MeshGeometry]{}, err
	}

	return Identifiable[*graphics.MeshGeometry]{
		ID:    assetGeometry.ID,
		Value: meshGeometry,
	}, nil
}

func (l *AssetLoader) ResolveMeshGeometries(assetGeometries []dto.Geometry) (IdentifiableList[*graphics.MeshGeometry], error) {
	geometries := make(IdentifiableList[*graphics.MeshGeometry], len(assetGeometries))
	var group errgroup.Group
	for i, assetGeometry := range assetGeometries {
		group.Go(func() error {
			geometry, err := l.ResolveMeshGeomety(assetGeometry)
			geometries[i] = geometry
			return err
		})
	}
	return geometries, group.Wait()
}

func (l *AssetLoader) ResolveMeshDefinition(assetDefinition dto.MeshDefinition, geometries IdentifiableList[*graphics.MeshGeometry], materials IdentifiableList[*graphics.Material]) (Identifiable[*graphics.MeshDefinition], error) {
	geometry, ok := geometries.FindByID(assetDefinition.GeometryID)
	if !ok {
		return Identifiable[*graphics.MeshDefinition]{}, fmt.Errorf("mesh geometry with ID %d not found", assetDefinition.GeometryID)
	}

	bindingMaterials := make([]*graphics.Material, geometry.FragmentCount())
	for _, assetBinding := range assetDefinition.MaterialBindings {
		material, ok := materials.FindByID(assetBinding.MaterialID)
		if !ok {
			return Identifiable[*graphics.MeshDefinition]{}, fmt.Errorf("material with ID %d not found", assetBinding.MaterialID)
		}
		bindingMaterials[assetBinding.FragmentIndex] = material
	}

	meshDefinitionInfo := graphics.MeshDefinitionInfo{
		Geometry:  geometry,
		Materials: bindingMaterials,
	}

	var meshDefinition *graphics.MeshDefinition
	allocateDefinition := func(engine *Engine) error {
		gfxEngine := engine.Graphics()
		meshDefinition = gfxEngine.CreateMeshDefinition(meshDefinitionInfo)
		return nil
	}
	if err := l.ScheduleMain(allocateDefinition).Wait(); err != nil {
		return Identifiable[*graphics.MeshDefinition]{}, err
	}

	return Identifiable[*graphics.MeshDefinition]{
		ID:    assetDefinition.ID,
		Value: meshDefinition,
	}, nil
}

func (l *AssetLoader) ResolveMeshDefinitions(assetDefinitions []dto.MeshDefinition, geometries IdentifiableList[*graphics.MeshGeometry], materials IdentifiableList[*graphics.Material]) (IdentifiableList[*graphics.MeshDefinition], error) {
	definitions := make(IdentifiableList[*graphics.MeshDefinition], len(assetDefinitions))
	var group errgroup.Group
	for i, assetDefinition := range assetDefinitions {
		group.Go(func() error {
			definition, err := l.ResolveMeshDefinition(assetDefinition, geometries, materials)
			definitions[i] = definition
			return err
		})
	}
	return definitions, group.Wait()
}

type MeshTemplate struct {
	NodeID       uint32
	DefinitionID uint32
	ArmatureID   uint32
}

func (l *AssetLoader) ResolveMeshTemplate(assetMesh dto.Mesh) (Identifiable[MeshTemplate], error) {
	return Identifiable[MeshTemplate]{
		ID: assetMesh.ID,
		Value: MeshTemplate{
			NodeID:       assetMesh.NodeID,
			DefinitionID: assetMesh.MeshDefinitionID,
			ArmatureID:   assetMesh.ArmatureID,
		},
	}, nil
}

func (l *AssetLoader) ResolveMeshTemplates(assetMeshes []dto.Mesh) (IdentifiableList[MeshTemplate], error) {
	templates := make(IdentifiableList[MeshTemplate], len(assetMeshes))
	for i, assetMesh := range assetMeshes {
		template, err := l.ResolveMeshTemplate(assetMesh)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

func (s *Scene) InstantiateMeshTemplateStatic(template MeshTemplate, nodes IdentifiableList[*hierarchy.Node], definitions IdentifiableList[*graphics.MeshDefinition], armatures IdentifiableList[*graphics.Armature]) {
	node := nodes.GetByID(template.NodeID)
	meshDefinition := definitions.GetByID(template.DefinitionID)
	var armature *graphics.Armature
	if template.ArmatureID != UnspecifiedID {
		armature = armatures.GetByID(template.ArmatureID)
	}
	s.gfxScene.CreateStaticMesh(graphics.StaticMeshInfo{
		Definition: meshDefinition,
		Armature:   armature,
		Matrix:     node.AbsoluteMatrix(),
	})
}

func (s *Scene) InstantiateMeshTemplateDynamic(template MeshTemplate, nodes IdentifiableList[*hierarchy.Node], definitions IdentifiableList[*graphics.MeshDefinition], armatures IdentifiableList[*graphics.Armature]) *graphics.Mesh {
	node := nodes.GetByID(template.NodeID)
	meshDefinition := definitions.GetByID(template.DefinitionID)
	var armature *graphics.Armature
	if template.ArmatureID != UnspecifiedID {
		armature = armatures.GetByID(template.ArmatureID)
	}
	mesh := s.gfxScene.CreateMesh(graphics.MeshInfo{
		Definition: meshDefinition,
		Armature:   armature,
	})
	mesh.SetMatrix(node.AbsoluteMatrix())
	node.SetTarget(MeshNodeTarget{
		Mesh: mesh,
	})
	return mesh
}

func (l *AssetLoader) resolveTopology(primitive dto.Topology) render.Topology {
	switch primitive {
	case dto.TopologyPoints:
		return render.TopologyPoints
	case dto.TopologyLineList:
		return render.TopologyLineList
	case dto.TopologyLineStrip:
		return render.TopologyLineStrip
	case dto.TopologyTriangleList:
		return render.TopologyTriangleList
	case dto.TopologyTriangleStrip:
		return render.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

func (l *AssetLoader) resolveVertexFormat(layout dto.VertexLayout) graphics.MeshGeometryVertexFormat {
	var result graphics.MeshGeometryVertexFormat
	if attrib := layout.Coord; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Coord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Normal; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Normal = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Tangent; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Tangent = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.TexCoord; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.TexCoord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Color; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Color = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Weights; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Weights = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Joints; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Joints = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      l.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	return result
}

func (l *AssetLoader) resolveVertexAttributeFormat(format dto.VertexAttributeFormat) render.VertexAttributeFormat {
	switch format {
	case dto.VertexAttributeFormatRGBA32F:
		return render.VertexAttributeFormatRGBA32F
	case dto.VertexAttributeFormatRGB32F:
		return render.VertexAttributeFormatRGB32F
	case dto.VertexAttributeFormatRG32F:
		return render.VertexAttributeFormatRG32F
	case dto.VertexAttributeFormatR32F:
		return render.VertexAttributeFormatR32F

	case dto.VertexAttributeFormatRGBA16F:
		return render.VertexAttributeFormatRGBA16F
	case dto.VertexAttributeFormatRGB16F:
		return render.VertexAttributeFormatRGB16F
	case dto.VertexAttributeFormatRG16F:
		return render.VertexAttributeFormatRG16F
	case dto.VertexAttributeFormatR16F:
		return render.VertexAttributeFormatR16F

	case dto.VertexAttributeFormatRGBA16S:
		return render.VertexAttributeFormatRGBA16S
	case dto.VertexAttributeFormatRGB16S:
		return render.VertexAttributeFormatRGB16S
	case dto.VertexAttributeFormatRG16S:
		return render.VertexAttributeFormatRG16S
	case dto.VertexAttributeFormatR16S:
		return render.VertexAttributeFormatR16S

	case dto.VertexAttributeFormatRGBA16SN:
		return render.VertexAttributeFormatRGBA16SN
	case dto.VertexAttributeFormatRGB16SN:
		return render.VertexAttributeFormatRGB16SN
	case dto.VertexAttributeFormatRG16SN:
		return render.VertexAttributeFormatRG16SN
	case dto.VertexAttributeFormatR16SN:
		return render.VertexAttributeFormatR16SN

	case dto.VertexAttributeFormatRGBA16U:
		return render.VertexAttributeFormatRGBA16U
	case dto.VertexAttributeFormatRGB16U:
		return render.VertexAttributeFormatRGB16U
	case dto.VertexAttributeFormatRG16U:
		return render.VertexAttributeFormatRG16U
	case dto.VertexAttributeFormatR16U:
		return render.VertexAttributeFormatR16U

	case dto.VertexAttributeFormatRGBA16UN:
		return render.VertexAttributeFormatRGBA16UN
	case dto.VertexAttributeFormatRGB16UN:
		return render.VertexAttributeFormatRGB16UN
	case dto.VertexAttributeFormatRG16UN:
		return render.VertexAttributeFormatRG16UN
	case dto.VertexAttributeFormatR16UN:
		return render.VertexAttributeFormatR16UN

	case dto.VertexAttributeFormatRGBA8S:
		return render.VertexAttributeFormatRGBA8S
	case dto.VertexAttributeFormatRGB8S:
		return render.VertexAttributeFormatRGB8S
	case dto.VertexAttributeFormatRG8S:
		return render.VertexAttributeFormatRG8S
	case dto.VertexAttributeFormatR8S:
		return render.VertexAttributeFormatR8S

	case dto.VertexAttributeFormatRGBA8SN:
		return render.VertexAttributeFormatRGBA8SN
	case dto.VertexAttributeFormatRGB8SN:
		return render.VertexAttributeFormatRGB8SN
	case dto.VertexAttributeFormatRG8SN:
		return render.VertexAttributeFormatRG8SN
	case dto.VertexAttributeFormatR8SN:
		return render.VertexAttributeFormatR8SN

	case dto.VertexAttributeFormatRGBA8U:
		return render.VertexAttributeFormatRGBA8U
	case dto.VertexAttributeFormatRGB8U:
		return render.VertexAttributeFormatRGB8U
	case dto.VertexAttributeFormatRG8U:
		return render.VertexAttributeFormatRG8U
	case dto.VertexAttributeFormatR8U:
		return render.VertexAttributeFormatR8U

	case dto.VertexAttributeFormatRGBA8UN:
		return render.VertexAttributeFormatRGBA8UN
	case dto.VertexAttributeFormatRGB8UN:
		return render.VertexAttributeFormatRGB8UN
	case dto.VertexAttributeFormatRG8UN:
		return render.VertexAttributeFormatRG8UN
	case dto.VertexAttributeFormatR8UN:
		return render.VertexAttributeFormatR8UN

	case dto.VertexAttributeFormatRGBA8IU:
		return render.VertexAttributeFormatRGBA8IU
	case dto.VertexAttributeFormatRGB8IU:
		return render.VertexAttributeFormatRGB8IU
	case dto.VertexAttributeFormatRG8IU:
		return render.VertexAttributeFormatRG8IU
	case dto.VertexAttributeFormatR8IU:
		return render.VertexAttributeFormatR8IU

	default:
		panic(fmt.Errorf("unsupported vertex attribute format: %d", format))
	}
}

func (l *AssetLoader) resolveIndexFormat(layout dto.IndexLayout) render.IndexFormat {
	switch layout {
	case dto.IndexLayoutUint16:
		return render.IndexFormatUnsignedU16
	case dto.IndexLayoutUint32:
		return render.IndexFormatUnsignedU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}
