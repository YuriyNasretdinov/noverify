package linter

import (
	"bytes"
	"errors"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/VKCOM/noverify/src/lintdebug"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/php7"
	"github.com/VKCOM/noverify/src/rules"
)

type FileInfo struct {
	Filename string
	Contents []byte
}

func isPHPExtension(filename string) bool {
	fileExt := filepath.Ext(filename)
	if fileExt == "" {
		return false
	}

	fileExt = fileExt[1:] // cut "." in the beginning

	for _, ext := range PHPExtensions {
		if fileExt == ext {
			return true
		}
	}

	return false
}

func makePHPExtensionSuffixes() [][]byte {
	res := make([][]byte, 0, len(PHPExtensions))
	for _, ext := range PHPExtensions {
		res = append(res, []byte("."+ext))
	}
	return res
}

func isPHPExtensionBytes(filename []byte, suffixes [][]byte) bool {
	for _, suffix := range suffixes {
		if bytes.HasSuffix(filename, suffix) {
			return true
		}
	}

	return false
}

type ReadCallback func(ch chan FileInfo)

// ParseContents parses specified contents (or file) and returns *RootWalker.
// Function does not update global meta.
func ParseContents(filename string, contents []byte) (rootNode node.Node, w *RootWalker, err error) {
	parser := php7.NewParser(bytes.NewReader(contents), filename)
	parser.WithFreeFloating()
	parser.Parse()

	bufCopy := contents

	return analyzeFile(filename, bufCopy, parser)
}

func cloneRulesForFile(filename string, ruleSet *rules.ScopedSet) *rules.ScopedSet {
	if ruleSet == nil {
		return nil
	}

	var clone rules.ScopedSet
	for i, list := range &ruleSet.RulesByKind {
		res := make([]rules.Rule, 0, len(list))
		for _, rule := range list {
			if !strings.Contains(filename, rule.Path) {
				continue
			}
			ruleClone := rule
			ruleClone.Matcher = rule.Matcher.Clone()
			res = append(res, ruleClone)
		}
		clone.RulesByKind[i] = res
	}
	return &clone
}

func analyzeFile(filename string, contents []byte, parser *php7.Parser) (*node.Root, *RootWalker, error) {
	rootNode := parser.GetRootNode()

	if rootNode == nil {
		lintdebug.Send("Could not parse %s at all due to errors", filename)
		return nil, nil, errors.New("Empty root node")
	}

	w := &RootWalker{
		filename: filename,
		st:       &meta.ClassParseState{},

		// We need to clone rules since phpgrep matchers
		// contain mutable state that we don't want to share
		// between goroutines.
		anyRset:   cloneRulesForFile(filename, Rules.Any),
		rootRset:  cloneRulesForFile(filename, Rules.Root),
		localRset: cloneRulesForFile(filename, Rules.Local),
	}

	w.InitFromParser(contents, parser)
	w.InitCustom()

	rootNode.Walk(w)
	if meta.IsIndexingComplete() {
		AnalyzeFileRootLevel(rootNode, w)
	}
	for _, c := range w.custom {
		c.AfterLeaveFile()
	}

	for _, e := range parser.GetErrors() {
		w.Report(nil, LevelError, "syntax", "Syntax error: "+e.String())
	}

	return rootNode, w, nil
}

// AnalyzeFileRootLevel does analyze file top-level code.
// This method is exposed for language server use, you usually
// do not need to call it yourself.
func AnalyzeFileRootLevel(rootNode node.Node, d *RootWalker) {
	sc := meta.NewScope()
	sc.AddVarName("argv", meta.NewTypesMap("string[]"), "predefined", true)
	sc.AddVarName("argc", meta.NewTypesMap("int"), "predefined", true)
	b := &BlockWalker{
		ctx:                  &blockContext{sc: sc},
		r:                    d,
		unusedVars:           make(map[string][]node.Node),
		nonLocalVars:         make(map[string]struct{}),
		ignoreFunctionBodies: true,
		rootLevel:            true,
	}

	for _, createFn := range d.customBlock {
		b.custom = append(b.custom, createFn(&BlockContext{w: b}))
	}

	rootNode.Walk(b)
}

var bytesBufPool = sync.Pool{
	New: func() interface{} { return &bytes.Buffer{} },
}

// DebugMessage is used to actually print debug messages.
func DebugMessage(msg string, args ...interface{}) {
	if Debug {
		log.Printf(msg, args...)
	}
}

// ParseFilenames is used to do initial parsing of files.
func ParseFilenames(readFileNamesFunc ReadCallback) []*Report {
	start := time.Now()
	defer func() {
		lintdebug.Send("Processing time: %s", time.Since(start))

		meta.Info.Lock()
		defer meta.Info.Unlock()

		lintdebug.Send("Funcs: %d, consts: %d, files: %d", meta.Info.NumFunctions(), meta.Info.NumConstants(), meta.Info.NumFilesWithFunctions())
	}()

	needReports := meta.IsIndexingComplete()

	lintdebug.Send("Parsing using %d cores", MaxConcurrency)

	filenamesCh := make(chan FileInfo, 512)

	go func() {
		readFileNamesFunc(filenamesCh)
		close(filenamesCh)
	}()

	var allReports []*Report
	for f := range filenamesCh {
		allReports = append(allReports, doParseFile(f, needReports)...)
	}

	return allReports
}

func doParseFile(f FileInfo, needReports bool) (reports []*Report) {
	var err error

	if DebugParseDuration > 0 {
		start := time.Now()
		defer func() {
			if dur := time.Since(start); dur > DebugParseDuration {
				log.Printf("Parsing of %s took %s", f.Filename, dur)
			}
		}()
	}

	if needReports {
		var w *RootWalker
		_, w, err = ParseContents(f.Filename, f.Contents)
		if err == nil {
			reports = w.GetReports()
		}
	} else {
		err = IndexFile(f.Filename, f.Contents)
	}

	if err != nil {
		log.Printf("Failed parsing %s: %s", f.Filename, err.Error())
		lintdebug.Send("Failed parsing %s: %s", f.Filename, err.Error())
	}

	return reports
}

// InitStubs parses directory with PHPStorm stubs which has all internal PHP classes and functions declared.
func InitStubs() {
	meta.Info.InitStubs()
}
