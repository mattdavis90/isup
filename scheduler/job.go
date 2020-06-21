package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"isup/jobparser"
)

type Job struct {
	Interval *time.Duration
	Ok       *string
	Tests    map[string]*Test
	Alerters []string
}

func (j *Job) Check(validAlerters []string) error {
	if j.Interval == nil {
		interval := 30 * time.Second // Default is 30s
		j.Interval = &interval
	}
	if j.Ok == nil {
		return errors.New("Job.Ok cannot be empty")
	}
	for _, t := range j.Tests {
		err := t.Check()
		if err != nil {
			return err
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
			return errors.New(fmt.Sprintf("'%s' Alerter was not found", a))
		}
	}
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
			Bool("result", v).
			Err(err).
			Msg("Job finished")
		alerts <- Alert{
			Job:      name,
			State:    v,
			Alerters: j.Alerters,
		}

		select {
		case <-time.After(*j.Interval):
		case <-ctx.Done():
			return
		}
	}
}

func (j *Job) run(jobName string, ctx context.Context) (bool, error) {
	var wg sync.WaitGroup
	resC := make(chan jobparser.Value)

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
				Bool("result", v).
				Err(err).
				Msg("Test finished")

			resC <- jobparser.Value{
				Name:   name,
				Result: v,
			}
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
		res[r.Name] = r.Result
	}

	tree, err := jobparser.Parse("parser", []byte(*j.Ok))
	if err != nil {
		return false, err
	}

	iface, ok := tree.([]interface{})
	if !ok {
		return false, err
	} else {
		expr := iface[0].(jobparser.Evaluatable)

		return expr.Evaluate(&res)
	}
}
