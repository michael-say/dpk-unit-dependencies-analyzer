/*Copyright (c) 2020 Michael Saigachenko*/
package main

import (
	"os"

	dpk "github.com/michael-say/dpk-unit-dependencies-analyzer/dpk"
	gc "github.com/untillpro/gochips"
)

func main() {
	gc.Info("Delphi DPK unit dependency analyzer")
	gc.ExitIfFalse(len(os.Args) == 2, "Syntax: depanalyzer <DPK_FILE_PATH>")
	dpk.ParseDpk(os.Args[1])
}
