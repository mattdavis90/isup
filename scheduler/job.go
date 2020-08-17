package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"isup/jobparser"
)

type Job struct {
	Interval      *time.Duration
	PendingAfter  int `yaml:"pending_after"`
	AlertingAfter int `yaml:"alerting_after"`
	OkAfter       int `yaml:"ok_after"`
	Ok            *string
	Tests         map[string]*Test
	Alerters      []string
	Values        map[string]string
	ok            jobparser.Evaluatable
	state         State
	timeAtState   int
}

func (j *Job) Check(validAlerters []string) error {
	if j.Interval == nil {
		interval := 30 * time.Second // Default is 30s
		j.Interval = &interval
	}

	if j.Ok == nil {
		return fmt.Errorf("Job.Ok cannot be empty")
	}
	tree, err := jobparser.Parse("parser", []byte(*j.Ok))
	if err != nil {
		return fmt.Errorf("Job.Ok Error: %w", err)
	}
	iface, ok := tree.([]interface{})
	if !ok {
		return fmt.Errorf("Job.Ok Error: %w", err)
	} else {
		j.ok = iface[0].(jobparser.Evaluatable)
	}

	if j.PendingAfter < 0 {
		j.PendingAfter = 0
	}
	if j.AlertingAfter < 1 {
		j.AlertingAfter = 1
	}
	if j.OkAfter < 1 {
		j.OkAfter = 1
	}

	for n, t := range j.Tests {
		err := t.Check()
		if err != nil {
			return fmt.Errorf("Test: '%s' %w", n, err)
		}
	}
	for _, a := range j.Alerters {
		found := false
		for _, va := range validAlerters {
			if a == va {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("'%s' Alerter was not found", a)
		}
	}
	j.state = NoDataState
	j.timeAtState = 1
	return nil
}

func (j *Job) Run(name string, ctx context.Context, alerts chan Alert) {
	log.Info().
		Str("job", name).
		Int("no_tests", len(j.Tests)).
		Msg("Loading job")

	for {
		log.Info().
			Str("job", name).
			Msg("Job starting")
		v, err := j.run(name, ctx)
		log.Info().
			Str("job", name).
			Str("result", v.String()).
			Err(err).
			Msg("Job finished")
		alerts <- Alert{
			Job:      name,
			State:    v,
			Alerters: j.Alerters,
			Values:   j.Values,
		}

		select {
		case <-time.After(*j.Interval):
		case <-ctx.Done():
			return
		}
	}
}

func (j *Job) run(jobName string, ctx context.Context) (State, error) {
	var wg sync.WaitGroup
	resC := make(chan struct {
		string
		State
	})

	for n, t := range j.Tests {
		wg.Add(1)
		go func(name string, test Test) {
			defer wg.Done()

			log.Info().
				Str("job", jobName).
				Str("test", name).
				Msg("Test starting")
			v, err := test.Run(ctx)
			log.Info().
				Str("job", jobName).
				Str("test", name).
				Str("result", v.String()).
				Err(err).
				Msg("Test finished")

			resC <- struct {
				string
				State
			}{name, v}
		}(n, *t)
	}

	// Wait for all results and close the channel
	go func() {
		wg.Wait()
		close(resC)
	}()

	// Collect results and wait for channel close
	res := jobparser.Values{}
	for r := range resC {
		if r.State == NoDataState {
			j.state = NoDataState
			j.timeAtState = 0
			return NoDataState, nil
		}
		res[r.string] = r.State == OkState
	}

	ok, err := j.ok.Evaluate(&res)

	switch j.state {
	case NoDataState:
		if ok {
			if j.timeAtState >= j.OkAfter {
				j.state = OkState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		} else {
			if j.PendingAfter != 0 {
				j.state = PendingState
				j.timeAtState = 1
			} else if j.timeAtState >= j.AlertingAfter {
				j.state = AlertingState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		}
	case OkState:
		if !ok {
			if j.PendingAfter != 0 {
				j.state = PendingState
				j.timeAtState = 1
			} else if j.timeAtState >= j.AlertingAfter {
				j.state = AlertingState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		}
	case PendingState:
		if ok {
			if j.timeAtState >= j.OkAfter {
				j.state = OkState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		} else {
			if j.timeAtState >= j.AlertingAfter {
				j.state = AlertingState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		}
	case AlertingState:
		if ok {
			if j.timeAtState >= j.OkAfter {
				j.state = OkState
				j.timeAtState = 1
			} else {
				j.timeAtState += 1
			}
		}
	}

	return j.state, err
}
