package scheduler

import (
	"context"

	"github.com/rs/zerolog/log"
)

type Alert struct {
	Job      string
	State    bool
	Alerters []string
}

type Alerter struct {
	Default    *bool
	AlwaysSend *bool
	Request    *Request
}

func (a *Alerter) Check() error {
	err := a.Request.Check()
	if err != nil {
		return err
	}

	if a.Default == nil {
		def := false
		a.Default = &def
	}

	if a.AlwaysSend == nil {
		alwaysSend := false
		a.AlwaysSend = &alwaysSend
	}

	return nil
}

func (a *Alerter) Run(name string, ctx context.Context, alerts chan Alert) {
	state := make(map[string]bool, 0)

	for {
		select {
		case alert := <-alerts:
			prevState, exist := state[alert.Job]
			if !exist || prevState != alert.State || *a.AlwaysSend {
				state[alert.Job] = alert.State

				repl := Replacement{}
				repl = repl.WithEnv()
				repl["Job"] = alert.Job
				if alert.State {
					repl["State"] = "Ok"
				} else {
					repl["State"] = "Alerting"
				}

				resp, err := a.Request.Run(ctx, &repl)
				if err != nil || resp.StatusCode < 200 || resp.StatusCode > 200 {
					log.Warn().
						Str("alerter", name).
						Str("job", alert.Job).
						Int("status_code", resp.StatusCode).
						Err(err).
						Msg("Alert failed to send")
				} else {
					log.Info().
						Str("alerter", name).
						Str("job", alert.Job).
						Int("status_code", resp.StatusCode).
						Msg("Alert sent")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}