package std

import co "github.com/mokiat/lacking/ui/component"

// ElementData is the struct that should be used when configuring
// an Element component's data.
type ElementData = co.ElementData

// Element represents the most basic component, which is translated
// to a UI Element. All higher-order components eventually boil down to an
// Element.
var Element = co.Element
