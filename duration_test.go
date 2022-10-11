package iso8601_test

import (
	"testing"
	"time"

	"gitoa.ru/go-4devs/iso8601"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val    time.Duration
		expect string
	}{
		"1 day": {
			val:    time.Hour * 24,
			expect: "P1D",
		},
		"1 hour": {
			val:    time.Hour,
			expect: "PT1H",
		},
		"1 second": {
			val:    time.Second,
			expect: "PT1S",
		},
		"1 nanosecond": {
			val:    time.Nanosecond,
			expect: "PT0.000000001S",
		},
		"negative": {
			val:    -time.Hour * 24,
			expect: "-P1D",
		},
		"zero": {
			val:    time.Duration(0),
			expect: "PT0S",
		},
	}

	for name, test := range cases {
		result := iso8601.FormatDuration(test.val)
		if result != test.expect {
			t.Errorf("test:%v got:%v, expect:%v", name, result, test.expect)
		}
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		opts   []iso8601.Option
		parse  string
		expect time.Duration
	}{
		"base": {
			parse:  "P3Y6M4DT12H30M17S",
			expect: parseDuration(t, "30780h30m17s"),
		},
		"base ofer time": {
			parse:  "P3Y6M4DT12H30M17S",
			expect: parseDuration(t, "30756h30m17s"),
			opts: []iso8601.Option{
				iso8601.From(func() time.Time {
					return parseTime(t, "2006-01-02T15:04:05Z")
				}),
			},
		},
		"base ofer time with delimiter": {
			parse:  "PT12H30.5M",
			expect: parseDuration(t, "12h30m30s"),
			opts: []iso8601.Option{
				iso8601.From(func() time.Time {
					return parseTime(t, "2006-01-02T15:04:05Z")
				}),
			},
		},
		"zero time": {
			parse:  "P3Y6M4DT12H30M17S",
			expect: parseDuration(t, "30732h30m17s"),
			opts: []iso8601.Option{
				iso8601.From(func() time.Time {
					return time.Time{}
				}),
			},
		},
		"only time": {
			parse:  "PT12H30M17S",
			expect: parseDuration(t, "12h30m17s"),
		},
		"time with days": {
			parse:  "P10DT12H30M17S",
			expect: parseDuration(t, "252h30m17s"),
		},
		"time with days with options": {
			parse:  "P10DT12H30M17S",
			expect: parseDuration(t, "252h30m17s"),
			opts: []iso8601.Option{
				iso8601.From(func() time.Time {
					return time.Time{}
				}),
			},
		},
		"one day": {
			parse:  "P1D",
			expect: time.Hour * 24,
		},
		"1 nanosecond": {
			parse:  "PT0.000000001S",
			expect: time.Nanosecond,
		},
	}

	for name, test := range cases {
		dur, err := iso8601.ParseDuration(test.parse, test.opts...)
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}

		if dur != test.expect {
			t.Errorf("test: %v expect:%v given:%v", name, test.expect, dur)
		}
	}
}

func parseDuration(t *testing.T, in string) time.Duration {
	t.Helper()

	duration, err := time.ParseDuration(in)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return duration
}

func parseTime(t *testing.T, in string) time.Time {
	t.Helper()

	duration, err := time.Parse(time.RFC3339, in)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	return duration
}
