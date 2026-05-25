package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithLocation(time.Local)),
	}
}

func (s *Scheduler) Start(cfg *Config) error {
	for _, routine := range cfg.Routines {
		if !routine.Enabled {
			continue
		}
		r := routine
		_, err := s.cron.AddFunc(r.Schedule, func() {
			execute(r)
		})
		if err != nil {
			return fmt.Errorf("invalid schedule for routine %q: %w", r.Name, err)
		}
	}
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

func (s *Scheduler) NextRun(routine Routine) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(routine.Schedule)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(time.Now()), nil
}

func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}
