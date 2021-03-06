package domain

import (
	"fmt"

	"github.com/ktye/iv/apl"
)

type SingleDomain interface {
	To(*apl.Apl, apl.Value) (apl.Value, bool)
	String(*apl.Apl) string
}

// propagate is used by a SingleDomain function for successive values to propagate
// testing to the child.
func propagate(a *apl.Apl, v apl.Value, child SingleDomain) (apl.Value, bool) {
	if child == nil {
		return v, true
	}
	return child.To(a, v)
}

// Both tests if both arguments satisfy the same domain.
func Both(same SingleDomain) apl.Domain {
	return both{same}
}

type both struct {
	same SingleDomain
}

func (b both) To(a *apl.Apl, L, R apl.Value) (apl.Value, apl.Value, bool) {
	if b.same == nil {
		// TODO: we could panic:
		// 	panic("both: child (same) is nil")
		// or allow any:
		return L, R, true
	}
	l, ok := b.same.To(a, L)
	if ok == false {
		return L, R, false
	}
	r, ok := b.same.To(a, R)
	if ok == false {
		return L, R, false
	}
	return l, r, true
}
func (b both) String(a *apl.Apl) string {
	if b.same == nil {
		return "any"
	}
	return "both " + b.same.String(a)
}

// Any tests if either the left or the right arguments satisfy the child domain.
func Any(child SingleDomain) apl.Domain {
	return any{child}
}

type any struct {
	child SingleDomain
}

func (b any) To(a *apl.Apl, L, R apl.Value) (apl.Value, apl.Value, bool) {
	l, ok := b.child.To(a, L)
	if ok == true {
		return l, R, true
	}
	r, ok := b.child.To(a, R)
	if ok == true {
		return L, r, true
	}
	return L, R, false
}
func (b any) String(a *apl.Apl) string {
	return "any " + b.child.String(a)
}

func Split(left, right SingleDomain) apl.Domain {
	return split{left, right}
}

type split struct {
	left, right SingleDomain
}

func (s split) To(a *apl.Apl, L, R apl.Value) (apl.Value, apl.Value, bool) {
	var ok bool
	l := L
	if s.left != nil {
		l, ok = s.left.To(a, L)
		if ok == false {
			return L, R, false
		}
	}
	r := R
	if s.right != nil {
		r, ok = s.right.To(a, R)
		if ok == false {
			return L, R, false
		}
	}
	return l, r, true
}
func (s split) String(a *apl.Apl) string {
	// TODO: if we use domain for both function and operators,
	// for operators, L and R should print as LO and RO.
	ls := "any"
	rs := "any"
	if s.left != nil {
		ls = s.left.String(a)
	}
	if s.right != nil {
		rs = s.right.String(a)
	}
	return fmt.Sprintf("L %s R %s", ls, rs)
}

// Monadic is a Domain for primitive functions, which checks if the left argument is nil,
// and the right argument satisfies right.
// For operators, use MonadicOp.
func Monadic(right SingleDomain) apl.Domain {
	return monadic{right}
}

type monadic struct {
	right SingleDomain
}

func (m monadic) To(a *apl.Apl, L, R apl.Value) (apl.Value, apl.Value, bool) {
	if L != nil {
		return L, R, false
	}
	if m.right == nil {
		return L, R, true
	}
	if r, ok := m.right.To(a, R); ok {
		return nil, r, true
	}
	return L, R, false
}
func (m monadic) IsDyadic() bool { return false }
func (m monadic) String(a *apl.Apl) string {
	if m.right == nil {
		return "R any"
	}
	return m.right.String(a)
}

// Dyadic is a Domain which checks if the left argument of a primitive function is not nil
// and the argments are within the child domain.
// For operators use DyadicOp.
func Dyadic(child apl.Domain) apl.Domain {
	return dyadic{child}
}

type dyadic struct {
	child apl.Domain
}

func (d dyadic) To(a *apl.Apl, L, R apl.Value) (apl.Value, apl.Value, bool) {
	if L == nil {
		return L, R, false
	}
	if d.child == nil {
		return L, R, true
	}
	return d.child.To(a, L, R)
}
func (d dyadic) IsDyadic() bool { return true }
func (d dyadic) String(a *apl.Apl) string {
	if d.child == nil {
		return "L any, R any"
	}
	return d.child.String(a)
}

func Not(child SingleDomain) SingleDomain {
	return not{child}
}

type not struct {
	child SingleDomain
}

func (n not) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	if n.child == nil {
		return V, false
	}
	v, ok := n.child.To(a, V)
	return v, !ok
}

func (n not) String(a *apl.Apl) string {
	if n.child == nil {
		return "never (Not(nil))" // This should not be done.
	}
	return "!" + n.child.String(a)
}

func Or(child1, child2 SingleDomain) SingleDomain {
	return or{child1, child2}
}

type or struct {
	child1, child2 SingleDomain
}

func (n or) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	if n.child1 == nil || n.child2 == nil {
		return V, false
	}
	if v, ok := n.child1.To(a, V); ok {
		return v, true
	} else if v, ok := n.child2.To(a, V); ok {
		return v, true
	}
	return V, false
}

func (n or) String(a *apl.Apl) string {
	return "(" + n.child1.String(a) + " or " + n.child2.String(a) + ")"
}
