package jobparser

import (
	"errors"
	"fmt"
)

type Value struct {
	Name   string
	Result bool
}

type Values map[string]bool

type Evaluatable interface {
	Evaluate(r *Values) (bool, error)
}

type Expression struct {
	Terms []Term
}

type Term struct {
	Operator  string
	Statement Evaluatable
}

type Variable struct {
	Name string
}

func (e Expression) Evaluate(r *Values) (bool, error) {
	var res bool
	var err error

	for _, v := range e.Terms {
		t, e := v.Statement.Evaluate(r)
		if e != nil {
			err = e
		}

		switch v.Operator {
		case "&&":
			res = res && t
		case "||":
			res = res || t
		default:
			res = t
		}
	}

	return res, err
}

func (v Variable) Evaluate(r *Values) (bool, error) {
	res, ok := (*r)[v.Name]
	if !ok {
		return false, errors.New(fmt.Sprintf("Test '%s' not found on Job", v.Name))
	}
	if res {
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("Test '%s' failed", v.Name))
}
