package apl

import (
	"fmt"
	"strings"
)

// Env is the environment of the current lambda function.
// It contains local variables and a pointer to the parent environment.
type env struct {
	parent *env
	vars   map[string]Value
}

// lambda is a function expression in braces {...}.
// It is also known under the term dynamic function or dfn.
type lambda struct {
	body guardList
}

func (λ *lambda) String(a *Apl) string {
	if λ.body == nil {
		return "{}"
	}
	return fmt.Sprintf("{%s}", λ.body.String(a))
}

func (λ *lambda) Eval(a *Apl) (Value, error) {
	return λ, nil
}

func (λ *lambda) Call(a *Apl, l, r Value) (Value, error) {
	if λ.body == nil {
		return EmptyArray{}, nil
	}

	e := env{
		vars:   make(map[string]Value),
		parent: a.env,
	}
	save := a.env
	a.env = &e
	defer func() { a.env = save }()

	e.vars["⍺"] = l
	e.vars["⍵"] = r
	return λ.body.Eval(a)
}

// guardList is the body of a lambda expression.
// It represents a list of guarded expressions.
type guardList []*guardExpr

func (l guardList) String(a *Apl) string {
	v := make([]string, len(l))
	for i, g := range l {
		v[i] = g.String(a)
	}
	return strings.Join(v, "⋄")
}

// Eval evaluates the guardList.
// It checks the condition of each guardExpr.
// Expressions are only evaluated, if the condition returns true or
// is nil.
// The function returns after the first evaluated expression, if it is
// not an assignment.
func (l guardList) Eval(a *Apl) (Value, error) {
	if len(l) == 0 {
		return EmptyArray{}, nil
	}
	var ret Value = EmptyArray{}
	for i, g := range l {
		isa := isAssignment(g.e)
		if g.cond == nil && i < len(l)-1 && isa == false {
			return nil, fmt.Errorf("λ contains non-reachable code")
		}
		if v, err := g.Eval(a); err != nil {
			return nil, err
		} else if v != nil {
			ret = v
			if isa == false {
				return ret, nil
			}
		}
	}
	return ret, nil
}

// guardExpr contains a guarded expression.
// It's expressions is evaluated if the condition returns true or is nil.
type guardExpr struct {
	cond expr
	e    expr
}

func (g *guardExpr) String(a *Apl) string {
	if g.cond == nil {
		return g.e.String(a)
	} else {
		return g.cond.String(a) + ":" + g.e.String(a)
	}
}

// Eval evaluates a guarded expression.
// If the condition exists, it is evaluated and must return a bool or a
// number convertable to boolean.
// If the condition is nil or returns true, the expression is evaluated,
// otherwise nil is returned and no error.
func (g *guardExpr) Eval(a *Apl) (Value, error) {
	if g.cond == nil {
		return g.e.Eval(a)
	}

	v, err := g.cond.Eval(a)
	if err != nil {
		return nil, err
	}
	b, isbool := v.(Bool)
	if isbool == false {
		if n, ok := v.(Number); ok {
			if nb, ok := a.Tower.ToBool(n); ok {
				b = nb
				isbool = true
			}
		}
	}
	if isbool == false {
		return nil, fmt.Errorf("λ condition does not return a bool: %s", b.String(a))
	}

	if b == false {
		return nil, nil
	} else {
		return g.e.Eval(a)
	}
}
