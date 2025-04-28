package glsl

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
)

//go:embed template/*.glsl
var templates embed.FS

type constructFunc func(name string, data any) string

func load(sources ...fs.FS) constructFunc {
	rootTemplate := template.New("root").Delims("/*", "*/")
	for _, source := range sources {
		rootTemplate = template.Must(rootTemplate.ParseFS(source, "*.glsl"))
	}

	buffer := new(bytes.Buffer)
	return func(name string, data any) string {
		tmpl := rootTemplate.Lookup(name)
		if tmpl == nil {
			panic(fmt.Errorf("template %q not found", name))
		}
		buffer.Reset()
		if err := tmpl.Execute(buffer, data); err != nil {
			panic(fmt.Errorf("template %q exec error: %w", name, err))
		}
		return buffer.String()
	}
}

func shaderDir(name string) fs.FS {
	subDir, err := fs.Sub(templates, name)
	if err != nil {
		panic(fmt.Errorf("error opening %q shaders: %w", name, err))
	}
	return subDir
}

var construct = load(shaderDir("template"))
