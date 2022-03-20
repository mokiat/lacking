package asset

type resourcesDTO struct {
	Resources []resourceDTO `yaml:"resources"`
}

type resourceDTO struct {
	GUID string `yaml:"guid"`
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type dependenciesDTO struct {
	Dependencies []dependencyDTO `yaml:"dependencies"`
}

type dependencyDTO struct {
	SourceGUID string `yaml:"source_guid"`
	TargetGUID string `yaml:"target_guid"`
}
