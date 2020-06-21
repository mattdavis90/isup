package scheduler

import (
	"context"

	"github.com/rs/zerolog/log"
)

type Schedule struct {
	Jobs map[string]*Job
}

func (s *Schedule) Check(validAlerters []string) error {
	for _, j := range s.Jobs {
		err := j.Check(validAlerters)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schedule) Run(ctx context.Context, alerts chan Alert) {
	log.Info().
		Int("no_jobs", len(s.Jobs)).
		Msg("Loading schedule")

	for n, j := range s.Jobs {
		go j.Run(n, ctx, alerts)
	}

	select {
	case <-ctx.Done():
	}
}
