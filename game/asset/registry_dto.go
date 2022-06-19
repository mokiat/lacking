package asset

type resourcesDTO struct {
	Resources    []resourceDTO   `yaml:"resources"`
	Dependencies []dependencyDTO `yaml:"dependencies"`
}

type resourceDTO struct {
	GUID string `yaml:"guid"`
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type dependencyDTO struct {
	SourceGUID string `yaml:"source_guid"`
	TargetGUID string `yaml:"target_guid"`
}
