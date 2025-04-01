package test

import (
	"fmt"
	"time"

	"go.uber.org/mock/gomock"
)

type timeMatcher struct {
	expected time.Time
	delta    time.Duration
}

func (m *timeMatcher) Matches(x any) bool {
	actual, ok := x.(time.Time)
	if !ok {
		return false
	}
	diff := actual.Sub(m.expected)
	return diff.Abs() <= m.delta
}

func (m *timeMatcher) String() string {
	return fmt.Sprintf("is within %v of %v", m.delta, m.expected)
}

func WithinDuration(t time.Time, delta time.Duration) gomock.Matcher {
	return &timeMatcher{expected: t, delta: delta}
}
