// Command textpipe is the program built from the textpipe workflow's
// source: it evaluates, plans and applies that workflow.
package main

import (
	"github.com/itsHabib/hack-workbench-typed/adapters/file"
	"github.com/itsHabib/hack-workbench-typed/adapters/transform"
	"github.com/itsHabib/hack-workbench-typed/cli"
	"github.com/itsHabib/hack-workbench-typed/engine"
	"github.com/itsHabib/hack-workbench-typed/workflows/textpipe"
)

func main() {
	cli.Program{
		Define:   textpipe.Define,
		Adapters: []engine.Adapter{file.Adapter{}, transform.Adapter{}},
	}.Main()
}
