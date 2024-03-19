package lsl

func Parse(source string) (*Shader, error) {
	// FIXME:
	return &Shader{
		Declarations: []Declaration{
			&UniformBlockDeclaration{
				Fields: []Field{
					{
						Name: "color",
						Type: TypeNameVec4,
					},
				},
			},

			&FunctionDeclaration{
				Name: "#fragment",
				Body: []Statement{
					&Assignment{
						Target: "#color",
						Expression: &Identifier{
							Name: "color",
						},
					},
				},
			},
		},
	}, nil
}
