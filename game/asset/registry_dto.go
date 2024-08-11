package asset

type registryDTO struct {
	Resources    []resourceDTO   `json:"resources"`
	Dependencies []dependencyDTO `json:"dependencies"`
}

type resourceDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PreviewData  []byte `json:"preview"`
	SourceDigest string `json:"source_digest"`
}

type dependencyDTO struct {
	TargetID string `json:"target_id"`
	SourceID string `json:"source_id"`
}
