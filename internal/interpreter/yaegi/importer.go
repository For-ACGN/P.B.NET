package yaegi

import (
	"go/build"
)

// Importer is used to process "import" in source code.
// Beacon can remote load script from Controller.
type Importer struct {
	Context *build.Context
}

// Import is
func (im *Importer) Import(pkg string) map[string]string {
	return nil
}
