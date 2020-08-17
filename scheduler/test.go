package scheduler

import (
	"context"
	"fmt"

	"isup/testparser"
)

type Test struct {
	Ok       *string
	Request  *Request
	Response *Response
	ok       testparser.Evaluatable
}

func (t *Test) Check() error {
	if t.Ok == nil {
		return fmt.Errorf("Test.Ok cannot be empty")
	}
	tree, err := testparser.Parse("parser", []byte(*t.Ok))
	if err != nil {
		return fmt.Errorf("Test.Ok Error: %w", err)
	}
	iface, ok := tree.([]interface{})
	if !ok {
		return fmt.Errorf("Test.Ok Error: %w", err)
	} else {
		t.ok = iface[0].(testparser.Evaluatable)
	}

	if t.Request == nil {
		return fmt.Errorf("Test.Request is required")
	}
	err = t.Request.Check()
	if err != nil {
		return err
	}
	if t.Response == nil {
		t.Response = &Response{}
	}
	return t.Response.Check()
}

func (t *Test) Run(ctx context.Context) (State, error) {
	rep := Replacement{}
	rep = rep.WithEnv()
	resp, err := t.Request.Run(ctx, &rep)
	if err != nil {
		return NoDataState, err
	}
	defer resp.Body.Close()

	res, err := t.Response.Run(ctx, resp)
	if err != nil {
		return NoDataState, err
	}

	ok, err := t.ok.Evaluate(res)
	if ok {
		return OkState, nil
	}
	return AlertingState, err
}
