package asset

type registryDTO struct {
	Resources    []resourceDTO   `json:"resources"`
	Dependencies []dependencyDTO `json:"dependencies"`
}

type resourceDTO struct {
	ID          string `json:"guid"`
	Name        string `json:"name"`
	PreviewData []byte `json:"preview_data"`
}

type dependencyDTO struct {
	TargetID string `json:"target_id"`
	SourceID string `json:"source_id"`
}
