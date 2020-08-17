package testparser

import (
	"fmt"
)

type Value struct {
	StrValue string
	NumValue float64
}

type Values map[string]Value

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

type Comparison struct {
	Variable   string
	Comparator string
	IsString   bool
	StrValue   string
	NumValue   float64
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

func (c Comparison) Evaluate(r *Values) (bool, error) {
	v, ok := (*r)[c.Variable]
	if !ok {
		return false, fmt.Errorf("Var '%s' not found on test", c.Variable)
	}

	var ret bool
	var err error

	if c.IsString {
		switch c.Comparator {
		case "==":
			ret = v.StrValue == c.StrValue
			err = fmt.Errorf("Test Failed: %s(%s) == %s", c.Variable, v.StrValue, c.StrValue)
		case "!=":
			ret = v.StrValue != c.StrValue
			err = fmt.Errorf("Test Failed: %s(%s) != %s", c.Variable, v.StrValue, c.StrValue)
		default:
			ret = false
			err = fmt.Errorf("Bad Comparator")
		}
	} else {
		switch c.Comparator {
		case "==":
			ret = v.NumValue == c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) == %f", c.Variable, v.NumValue, c.NumValue)
		case "!=":
			ret = v.NumValue != c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) != %f", c.Variable, v.NumValue, c.NumValue)
		case ">":
			ret = v.NumValue > c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) > %f", c.Variable, v.NumValue, c.NumValue)
		case ">=":
			ret = v.NumValue >= c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) >= %f", c.Variable, v.NumValue, c.NumValue)
		case "<":
			ret = v.NumValue < c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) < %f", c.Variable, v.NumValue, c.NumValue)
		case "<=":
			ret = v.NumValue <= c.NumValue
			err = fmt.Errorf("Test Failed: %s(%f) <= %f", c.Variable, v.NumValue, c.NumValue)
		default:
			ret = false
			err = fmt.Errorf("Bad Comparator")
		}
	}

	if !ret {
		return false, err
	}
	return true, nil
}
