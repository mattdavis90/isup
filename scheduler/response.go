package scheduler

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"

	"isup/testparser"
)

type Response struct {
	Extract map[string]string
}

func (r *Response) Check() error {
	return nil
}

func (r *Response) Run(ctx context.Context, resp *http.Response) (*testparser.Values, error) {
	values := testparser.Values{
		"status_code": testparser.Value{
			NumValue: float64(resp.StatusCode),
		},
	}

	if len(r.Extract) > 0 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if gjson.Valid(string(data)) {
			json := gjson.Parse(string(data))
			for n, p := range r.Extract {
				var r testparser.Value
				v := json.Get(p)

				switch v.Type {
				case gjson.String:
					r = testparser.Value{
						StrValue: v.Str,
					}
				case gjson.Number:
					r = testparser.Value{
						NumValue: v.Num,
					}
				}
				values[n] = r

			}
		} else {
			return nil, errors.New("Unable to parse JSON")
		}
	}

	return &values, nil
}
