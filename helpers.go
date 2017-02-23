package plush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/gobuffalo/plush/ast"

	"github.com/markbates/inflect"
	"github.com/pkg/errors"
)

// Helpers contains all of the default helpers for
// These will be available to all templates. You should add
// any custom global helpers to this list.
var Helpers = HelperMap{}

func init() {
	Helpers.Add("json", toJSONHelper)
	Helpers.Add("jsEscape", template.JSEscapeString)
	Helpers.Add("htmlEscape", template.HTMLEscapeString)
	Helpers.Add("upcase", strings.ToUpper)
	Helpers.Add("downcase", strings.ToLower)
	Helpers.Add("contentFor", contentForHelper)
	Helpers.Add("contentOf", contentOfHelper)
	Helpers.Add("markdown", markdownHelper)
	Helpers.Add("len", lenHelper)
	Helpers.Add("debug", debugHelper)
	Helpers.Add("inspect", inspectHelper)
	Helpers.AddMany(inflect.Helpers)
}

// HelperContext is an optional last argument to helpers
// that provides the current context of the call, and access
// to an optional "block" of code that can be executed from
// within the helper.
type HelperContext struct {
	*Context
	ev    *evaler
	block *ast.BlockStatement
}

var helperContextKind = "HelperContext"

// Block executes the block of template associated with
// the helper, think the block inside of an "if" or "each"
// statement.
func (h HelperContext) Block() (string, error) {
	return h.BlockWith(h.Context)
}

// BlockWith executes the block of template associated with
// the helper, think the block inside of an "if" or "each"
// statement, but with it's own context.
func (h HelperContext) BlockWith(ctx *Context) (string, error) {
	octx := h.ev.ctx
	defer func() { h.ev.ctx = octx }()
	h.ev.ctx = ctx

	if h.block == nil {
		return "", errors.New("no block defined")
	}
	i, err := h.ev.evalBlockStatement(h.block)
	if err != nil {
		return "", err
	}
	bb := &bytes.Buffer{}
	h.ev.write(bb, i)
	return bb.String(), nil
}

// Helpers associated with the current context.
func (h HelperContext) Helpers() *HelperMap {
	return &h.ev.template.Helpers
}

// toJSONHelper converts an interface into a string.
func toJSONHelper(v interface{}) (template.HTML, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return template.HTML(b), nil
}

func lenHelper(v interface{}) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Len()
}

// Debug by verbosely printing out using 'pre' tags.
func debugHelper(v interface{}) template.HTML {
	return template.HTML(fmt.Sprintf("<pre>%+v</pre>", v))
}

func inspectHelper(v interface{}) string {
	return fmt.Sprintf("%+v", v)
}