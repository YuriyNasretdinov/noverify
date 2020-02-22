package linter

import (
	"fmt"
	"strings"
)

const (
	// IgnoreLinterMessage is a commit message that you specify if you want to cancel linter checks for this changeset
	IgnoreLinterMessage = "@linter disable"
)

func init() {
	_ = []CheckInfo{
		{
			Name:    "discardExpr",
			Default: true,
			Comment: `Report expressions that are evaluated but not used.`,
		},

		{
			Name:    "voidResultUsed",
			Default: true,
			Comment: `Report usages of the void-type expressions`,
		},

		{
			Name:    "keywordCase",
			Default: true,
			Comment: `Report keywords that are not in the lower case.`,
		},

		{
			Name:    "accessLevel",
			Default: true,
			Comment: `Report erroneous member access.`,
		},

		{
			Name:    "argCount",
			Default: true,
			Comment: `Report mismatching args count inside call expressions.`,
		},

		{
			Name:    "arrayAccess",
			Default: true,
			Comment: `Report array access to non-array objects.`,
		},

		{
			Name:    "bitwiseOps",
			Default: true,
			Comment: `Report suspicious usage of bitwise operations.`,
		},

		{
			Name:    "mixedArrayKeys",
			Default: true,
			Comment: `Report array literals that have both implicit and explicit keys.`,
		},

		{
			Name:    "dupArrayKeys",
			Default: true,
			Comment: `Report duplicated keys in array literals.`,
		},

		{
			Name:    "arraySyntax",
			Default: true,
			Comment: `Report usages of old array() syntax.`,
		},

		{
			Name:    "bareTry",
			Default: true,
			Comment: `Report try blocks without catch/finally.`,
		},

		{
			Name:    "caseBreak",
			Default: true,
			Comment: `Report switch cases without break.`,
		},

		{
			Name:    "complexity",
			Default: true,
			Comment: `Report funcs/methods that are too complex.`,
		},

		{
			Name:    "deadCode",
			Default: true,
			Comment: `Report potentially unreachable code.`,
		},

		{
			Name:    "phpdocLint",
			Default: true,
			Comment: `Report malformed phpdoc comments.`,
		},

		{
			Name:    "phpdocType",
			Default: true,
			Comment: `Report potential issues in phpdoc types.`,
		},

		{
			Name:    "phpdoc",
			Default: true,
			Comment: `Report missing phpdoc on public methods.`,
		},

		{
			Name:    "stdInterface",
			Default: true,
			Comment: `Report issues related to std PHP interfaces.`,
		},

		{
			Name:    "syntax",
			Default: true,
			Comment: `Report syntax errors.`,
		},

		{
			Name:    "undefined",
			Default: true,
			Comment: `Report usages of potentially undefined symbols.`,
		},

		{
			Name:    "unused",
			Default: true,
			Comment: `Report potentially unused variables.`,
		},

		{
			Name:    "redundantCast",
			Default: false,
			Comment: `Report redundant type casts.`,
		},

		{
			Name:    "caseContinue",
			Default: true,
			Comment: `Report suspicious 'continue' usages inside switch cases.`,
		},

		{
			Name:    "deprecated",
			Default: false, // Experimental
			Comment: `Report usages of deprecated symbols.`,
		},

		{
			Name:    "callStatic",
			Default: true,
			Comment: `Report static calls of instance methods and vice versa.`,
		},

		{
			Name:    "oldStyleConstructor",
			Default: true,
			Comment: `Report old-style (PHP4) class constructors.`,
		},
	}

	// for _, info := range allChecks {
	// DeclareCheck(info)
	// }
}

// Report is a linter report message.
type Report struct {
	checkName  string
	startLn    string
	startChar  int
	startLine  int
	endChar    int
	level      int
	msg        string
	filename   string
	isDisabled bool // user-defined flag that file should not be linted
}

// CheckName returns report associated check name.
func (r *Report) CheckName() string {
	return r.checkName
}

func (r *Report) String() string {
	contextLn := strings.Builder{}
	for i, ch := range r.startLn {
		if i == r.startChar {
			break
		}
		if ch == '\t' {
			contextLn.WriteRune(ch)
		} else {
			contextLn.WriteByte(' ')
		}
	}

	if r.endChar > r.startChar {
		contextLn.WriteString(strings.Repeat("^", r.endChar-r.startChar))
	}

	msg := r.msg
	if r.checkName != "" {
		msg = r.checkName + ": " + msg
	}
	return fmt.Sprintf("%s %s at %s:%d\n%s\n%s", severityNames[r.level], msg, r.filename, r.startLine, r.startLn, contextLn.String())
}

// IsCritical returns whether or not we need to reject whole commit when found this kind of report.
func (r *Report) IsCritical() bool {
	return r.level != LevelDoNotReject
}

// IsDisabledByUser returns whether or not user thinks that this file should not be checked
func (r *Report) IsDisabledByUser() bool {
	return r.isDisabled
}

// GetFilename returns report filename
func (r *Report) GetFilename() string {
	return r.filename
}
