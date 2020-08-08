package phpgrep

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
)

type metaNode struct {
	name string
}

func (metaNode) Walk(v ir.Visitor)                         {}
func (metaNode) GetFreeFloating() *freefloating.Collection { return nil }

type (
	anyConst struct{ metaNode }
	anyVar   struct{ metaNode }
	anyInt   struct{ metaNode }
	anyFloat struct{ metaNode }
	anyStr   struct{ metaNode }
	anyNum   struct{ metaNode }
	anyExpr  struct{ metaNode }
	anyFunc  struct{ metaNode }
)
