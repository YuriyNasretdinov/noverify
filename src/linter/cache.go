package linter

import (
	"errors"

	"github.com/VKCOM/noverify/src/meta"
)

// cacheVersions is a magic number that helps to distinguish incompatible caches.
//
// Version log:
//     27 - added Static field to meta.FuncInfo
//     28 - array type parsed as mixed[]
//     29 - updated type inference for ClassConstFetch
//     30 - resolve ClassConstFetch to a wrapped type string
//     31 - fixed plus operator type inference for arrays
//     32 - replaced Static:bool with Flags:uint8 in meta.FuncInfo
//     33 - support parsing of array<k,v> and list<type>
//     34 - support parsing of ?ClassName as "ClassName|null"
const cacheVersion = 34

var (
	errWrongVersion = errors.New("Wrong cache version")

	initFileReadTime  int64
	initCacheReadTime int64
)

type fileMeta struct {
	Scope             *meta.Scope
	Classes           meta.ClassesMap
	Traits            meta.ClassesMap
	Functions         meta.FunctionsMap
	Constants         meta.ConstantsMap
	FunctionOverrides meta.FunctionsOverrideMap
}

// IndexFile parses the file and fills in the meta info. Can use cache.
func IndexFile(filename string, contents []byte) error {
	_, w, err := ParseContents(filename, contents)
	if w != nil {
		updateMetaInfo(filename, &w.meta)
	}
	return err
}

func updateMetaInfo(filename string, m *fileMeta) {
	if meta.IsIndexingComplete() {
		panic("Trying to update meta info when not indexing")
	}

	meta.Info.Lock()
	defer meta.Info.Unlock()

	meta.Info.DeleteMetaForFileNonLocked(filename)

	meta.Info.AddFilenameNonLocked(filename)
	meta.Info.AddClassesNonLocked(filename, m.Classes)
	meta.Info.AddTraitsNonLocked(filename, m.Traits)
	meta.Info.AddFunctionsNonLocked(filename, m.Functions)
	meta.Info.AddConstantsNonLocked(filename, m.Constants)
	meta.Info.AddFunctionsOverridesNonLocked(filename, m.FunctionOverrides)

	if m.Scope != nil {
		meta.Info.AddToGlobalScopeNonLocked(filename, m.Scope)
	}
}
