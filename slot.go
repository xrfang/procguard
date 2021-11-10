package procguard

import (
	"fmt"
	"sort"
	"time"
)

type (
	slot struct {
		Since string
		Until string
	}
	slots map[string][]slot
)

func (s slot) validate() error {
	ts, err := time.Parse("15:04:05", s.Since)
	if err != nil {
		return err
	}
	tu, err := time.Parse("15:04:05", s.Until)
	if err != nil {
		return err
	}
	if ts.After(tu) {
		return fmt.Errorf("since (%s) after until (%s)", s.Since, s.Until)
	}
	return nil
}

func (s slot) active(t string) bool {
	return s.Since <= t && s.Until >= t
}

func (ss slots) validate() error {
	for k, sl := range ss {
		for i, s := range sl {
			if err := s.validate(); err != nil {
				return fmt.Errorf("[%s]slot#%d: %v", k, i, err)
			}
		}
	}
	return nil
}

func (ss slots) current(now string) []string {
	tags := make(map[string]bool)
	for t, ls := range ss {
		for _, l := range ls {
			if l.active(now) {
				tags[t] = true
				break
			}
		}
	}
	var ts []string
	for t := range tags {
		ts = append(ts, t)
	}
	sort.Strings(ts)
	return ts
}
