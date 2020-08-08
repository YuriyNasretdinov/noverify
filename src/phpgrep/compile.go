package phpgrep

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/irgen"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

type compiler struct {
	src  []byte
	vars map[string]struct{}
}

func compile(opts *Compiler, pattern []byte) (*Matcher, error) {
	root, src, err := parseutil.Parse(pattern)
	if err != nil {
		return nil, err
	}
	rootIR := irgen.ConvertNode(root)

	if st, ok := rootIR.(*ir.ExpressionStmt); ok {
		rootIR = st.Expr
	}

	c := compiler{
		src:  src,
		vars: make(map[string]struct{}),
	}
	rootIR.Walk(&c)

	m := &Matcher{
		m: matcher{
			root:    rootIR,
			numVars: len(c.vars),
		},
	}

	return m, nil
}

func (c *compiler) EnterNode(n ir.Node) bool {
	if v, ok := n.(*ir.SimpleVar); ok {
		c.vars[v.Name] = struct{}{}
		return true
	}

	v, ok := n.(*ir.Var)
	if !ok {
		return true
	}
	s, ok := v.Expr.(*ir.String)
	if !ok {
		return true
	}
	value := unquoted(s.Value)

	var name string
	var class string

	colon := strings.Index(value, ":")
	if colon == -1 {
		// Anonymous matcher.
		name = "_"
		class = value
	} else {
		// Named matcher.
		name = value[:colon]
		class = value[colon+len(":"):]
		c.vars[name] = struct{}{}
	}

	switch class {
	case "var":
		v.Expr = anyVar{metaNode{name: name}}
	case "int":
		v.Expr = anyInt{metaNode{name: name}}
	case "float":
		v.Expr = anyFloat{metaNode{name: name}}
	case "str":
		v.Expr = anyStr{metaNode{name: name}}
	case "num":
		v.Expr = anyNum{metaNode{name: name}}
	case "expr":
		v.Expr = anyExpr{metaNode{name: name}}
	case "const":
		v.Expr = anyConst{metaNode{name: name}}
	case "func":
		v.Expr = anyFunc{metaNode{name: name}}
	}

	return true
}

func (c *compiler) LeaveNode(n ir.Node) {}
