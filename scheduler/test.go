package scheduler

import (
	"context"
	"errors"

	"isup/testparser"
)

type Test struct {
	Ok       *string
	Request  *Request
	Response *Response
}

func (t *Test) Check() error {
	if t.Ok == nil {
		return errors.New("Test.Ok cannot be empty")
	}
	if t.Request == nil {
		return errors.New("Test.Request is required")
	}
	err := t.Request.Check()
	if err != nil {
		return err
	}
	if t.Response == nil {
		t.Response = &Response{}
	}
	return t.Response.Check()
}

func (t *Test) Run(ctx context.Context) (bool, error) {
	rep := Replacement{}
	rep = rep.WithEnv()
	resp, err := t.Request.Run(ctx, &rep)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	res, err := t.Response.Run(ctx, resp)
	if err != nil {
		return false, err
	}

	tree, err := testparser.Parse("parser", []byte(*t.Ok))
	if err != nil {
		return false, err
	}

	iface, ok := tree.([]interface{})
	if !ok {
		return false, err
	} else {
		expr := iface[0].(testparser.Evaluatable)
		ok, err := expr.Evaluate(res)
		if ok {
			return true, nil
		}
		return false, err
	}
}
