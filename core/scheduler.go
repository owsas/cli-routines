package core

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler manages cron-based scheduling of routines.
type Scheduler struct {
	cron *cron.Cron
}

// NewScheduler creates a new Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithLocation(time.Local)),
	}
}

// Start begins scheduling all enabled routines.
func (s *Scheduler) Start(cfg *Config) error {
	for i := range cfg.Routines {
		if !cfg.Routines[i].Enabled {
			continue
		}
		r := &cfg.Routines[i]
		if err := r.Resolve(); err != nil {
			AppendLog(fmt.Sprintf("Skipping routine %q: cannot resolve executor: %v", r.Name, err))
			continue
		}
		_, err := s.cron.AddFunc(r.Schedule, func() {
			Execute(*r)
		})
		if err != nil {
			return fmt.Errorf("invalid schedule for routine %q: %w", r.Name, err)
		}
	}
	s.cron.Start()
	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}

// NextRun calculates the next time a routine will run.
func (s *Scheduler) NextRun(routine Routine) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	sched, err := parser.Parse(routine.Schedule)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(time.Now()), nil
}

// Entries returns the currently scheduled cron entries.
func (s *Scheduler) Entries() []cron.Entry {
	return s.cron.Entries()
}
