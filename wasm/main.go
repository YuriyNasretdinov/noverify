package main

import (
	"strings"
	"syscall/js"
	"time"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/vscode"
)

func parse(filename string, contents string) (rootNode node.Node, w *linter.RootWalker, err error) {
	rootNode, w, err = linter.ParseContents(filename, []byte(contents))
	if err != nil {
		return nil, nil, err
	}

	if !meta.IsIndexingComplete() {
		w.UpdateMetaInfo()
	}

	return rootNode, w, nil
}

func getReports(contents string) ([]vscode.Diagnostic, error) {
	meta.ResetInfo()
	if _, _, err := parse(`demo.php`, contents); err != nil {
		return nil, err
	}
	meta.SetIndexingComplete(true)
	_, w, err := parse(`demo.php`, contents)
	if err != nil {
		return nil, err
	}
	return w.Diagnostics, err
}

var needAnalyse = false

func doAnalyse() {
	text := js.Global().Get("editor").Call("getValue").String()
	diags, err := getReports(text)

	var value string
	if err != nil {
		value = "ERROR: " + err.Error()
	} else {
		var ds []string
		for _, d := range diags {
			ds = append(ds, d.Message)
		}
		value = strings.Join(ds, "\n")
	}

	js.Global().Call("showErrors", value)
}

func main() {
	linter.LangServer = true

	go func() {
		for {
			time.Sleep(time.Second)
			if needAnalyse {
				needAnalyse = false
				doAnalyse()
			}
		}
	}()

	js.Global().Set("analyzeCallback", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		needAnalyse = true
		return nil
	}))

	select {}
}
