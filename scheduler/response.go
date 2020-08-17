package scheduler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
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
		log.Debug().Str("body", string(data)).Msg("Response body")
		if gjson.Valid(string(data)) {
			json := gjson.Parse(string(data))
			for n, p := range r.Extract {
				var r testparser.Value
				v := json.Get(p)
				log.Debug().Str("name", n).Str("path", p).Interface("value", v).Msg("extraction")

				switch v.Type {
				case gjson.String:
					num, _ := strconv.ParseFloat(v.Str, 64)
					r = testparser.Value{
						StrValue: v.Str,
						NumValue: num,
					}
				case gjson.Number:
					str := strconv.FormatFloat(v.Num, 'g', -1, 64)
					r = testparser.Value{
						StrValue: str,
						NumValue: v.Num,
					}
				}

				log.Debug().Str("name", n).Str("path", p).Interface("result", r).Msg("extraction_result")
				values[n] = r

			}
		} else {
			return nil, fmt.Errorf("Unable to parse JSON")
		}
	}

	return &values, nil
}
