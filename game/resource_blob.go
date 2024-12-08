package game

import "github.com/mokiat/lacking/game/asset"

func (s *ResourceSet) convertBlob(assetBlob asset.Blob) blob {
	return blob{
		name: assetBlob.Name,
		data: assetBlob.Data,
	}
}
