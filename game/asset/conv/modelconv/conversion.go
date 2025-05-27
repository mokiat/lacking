package modelconv

import (
	// Importing the following for side effects.
	_ "github.com/mokiat/lacking/game/asset/conv/animationconv"
	_ "github.com/mokiat/lacking/game/asset/conv/backgroundconv"
	_ "github.com/mokiat/lacking/game/asset/conv/hierarchyconv"
	_ "github.com/mokiat/lacking/game/asset/conv/lightingconv"
	_ "github.com/mokiat/lacking/game/asset/conv/meshconv"
	_ "github.com/mokiat/lacking/game/asset/conv/physicsconv"
	_ "github.com/mokiat/lacking/game/asset/conv/shadingconv"
)

func Ensure() {
	// No-op function to ensure that the package is initialized.
}
