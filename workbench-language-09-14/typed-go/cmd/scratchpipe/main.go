// Command scratchpipe is the program built from the scratchpipe workflow:
// textpipe plus a disposable directory, served by the third adapter.
package main

import (
	"github.com/itsHabib/hack-workbench-typed/adapters/file"
	"github.com/itsHabib/hack-workbench-typed/adapters/scratchdir"
	"github.com/itsHabib/hack-workbench-typed/adapters/transform"
	"github.com/itsHabib/hack-workbench-typed/cli"
	"github.com/itsHabib/hack-workbench-typed/engine"
	"github.com/itsHabib/hack-workbench-typed/workflows/scratchpipe"
)

func main() {
	cli.Program{
		Define:   scratchpipe.Define,
		Adapters: []engine.Adapter{file.Adapter{}, transform.Adapter{}, scratchdir.Adapter{}},
	}.Main()
}
