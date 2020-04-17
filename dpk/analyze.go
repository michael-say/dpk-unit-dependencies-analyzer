/*
	Copyright (c) 2020 Michael Saigachenko
*/

package dpk

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
	gc "github.com/untillpro/gochips"
)

type module struct {
	name     string
	path     string
	usesIntf []*module
}

func parseDpkStr(text string) []*module {
	modules := make([]*module, 0)
	parser := createParser(text)
	parser.skipUntil("(?i)\\s*contains\\s*")
	for {
		line := parser.nextLine()
		re := regexp.MustCompile("\\s*([a-zA-Z0-9_]+)\\sin\\s'(.*)'.*,")
		submatch := re.FindStringSubmatch(line)
		if len(submatch) > 2 {
			modules = append(modules, &module{
				name: submatch[1],
				path: submatch[2],
			})
		}
		if parser.eof() {
			break
		}
	}
	gc.Info("Package parsed. Modules:", len(modules))
	return modules
}

func parseUses(usesStr string, modules []*module) []*module {
	usesStr = strings.Trim(usesStr, "\n\r\t ;")
	units := strings.Split(usesStr, ",")
	uses := make([]*module, 0)
	for i := range units {
		unit := strings.Trim(units[i], " \t")
		for j := range modules {
			if strings.EqualFold(unit, modules[j].name) {
				uses = append(uses, modules[j])
				break
			}
		}
	}
	return uses
}

func readUsesIntf(wd string, file string, modules []*module) []*module {
	path := filepath.Join(wd, file)
	gc.ExitIfFalse(fileExists(path), "Unit file not exists", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	gc.ExitIfError(err)
	content, err := ioutil.ReadAll(f)
	gc.ExitIfError(err)
	parser := createParser(string(content))

	items := make([]*module, 0)
	parser.skipUntil("implementation")
	//parser.skipUntil("interface")
	for {
		line := parser.nextLine()
		re := regexp.MustCompile("(?i)uses\\s(.*)")
		submatch := re.FindStringSubmatch(strings.ToLower(line))
		if len(submatch) > 1 {
			var str strings.Builder
			str.WriteString(submatch[1])
			for {
				if strings.Index(str.String(), ";") > -1 {
					break
				}
				gc.ExitIfFalse(!parser.eof(), "Unexpected EOF")
				str.WriteString(parser.nextLine())
			}
			items = parseUses(str.String(), modules)
			break
		}
		if parser.eof() || len(items) > 0 {
			break
		}
	}
	return items
}

func indexOf(chain *list.List, m *module) int {
	index := 0
	for chain := chain.Front(); chain != nil; chain = chain.Next() {
		if chain.Value == m {
			return index
		}
		index++
	}
	return -1
}

func chainToStr(chain *list.List, m *module) string {
	var b strings.Builder
	for chain := chain.Front(); chain != nil; chain = chain.Next() {
		mm := chain.Value.(*module)
		b.WriteString(mm.name)
		b.WriteString(" -> ")
	}
	b.WriteString(m.name)
	return b.String()
}

func analyzeModule(m *module, chain *list.List) (bool, string) {
	if indexOf(chain, m) > -1 {
		return false, "Cyclic reference found: " + chainToStr(chain, m)
	}
	e := chain.PushBack(m)
	for i := range m.usesIntf {
		ret, msg := analyzeModule(m.usesIntf[i], chain)
		if !ret {
			return false, msg
		}
	}
	chain.Remove(e)
	return true, ""
}

func analyze(wd string, modules []*module) {
	for i := range modules {
		modules[i].usesIntf = readUsesIntf(wd, modules[i].path, modules)
	}

	bar := progressbar.New(len(modules))
	messages := make([]string, 0)

	count := 0
	for i := range modules {
		chain := list.New()
		m := modules[i]
		ret, msg := analyzeModule(m, chain)
		if !ret {
			count++
			messages = append(messages, msg)
		}
		bar.Add(1)
	}

	gc.Info("\n\n")
	for _, msg := range messages {
		gc.Info(msg)
	}
	gc.ExitIfFalse(count == 0, fmt.Sprintf("\n\nErrors found: cyclic references (%d)", count))
	gc.Info("done")
}

//ParseDpk parses DPK
func ParseDpk(dpk string) {
	gc.ExitIfFalse(fileExists(dpk), "File not exists", dpk)
	gc.Info("Analyzing dpk: ", dpk)
	file, err := os.OpenFile(dpk, os.O_RDONLY, 0600)
	gc.ExitIfError(err)
	content, err := ioutil.ReadAll(file)
	gc.ExitIfError(err)
	modules := parseDpkStr(string(content))
	analyze(path.Dir(dpk), modules)
}
