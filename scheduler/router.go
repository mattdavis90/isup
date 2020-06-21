package scheduler

import (
	"context"
)

type Router struct {
	Alerts   chan Alert
	Alerters map[string]*Alerter
}

func NewRouter() *Router {
	alerts := make(chan Alert, 100)
	return &Router{
		Alerts: alerts,
	}
}

func (r *Router) Run(ctx context.Context) error {
	chans := make(map[string]chan Alert, len(r.Alerters))

	for n, a := range r.Alerters {
		c := make(chan Alert, 100)
		chans[n] = c
		go a.Run(n, ctx, c)
	}

	for {
		select {
		case alert := <-r.Alerts:
			// If no alerters then use default ones
			if alert.Alerters == nil {
				alert.Alerters = make([]string, 0, len(r.Alerters))
				for n, a := range r.Alerters {
					if *a.Default {
						alert.Alerters = append(alert.Alerters, n)
					}
				}
			}

			for _, a := range alert.Alerters {
				c := chans[a]
				c <- alert
			}

		case <-ctx.Done():
			return nil
		}
	}
}
