package iso8601

import (
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrInvalidDuration = errors.New("invalid duration")
	ErrMissingUnit     = errors.New("missing unit")
	ErrUnknownUnit     = errors.New("unknown unit")
	ErrOverflow        = errors.New("overflow")
	ErrLeadingInt      = errors.New("leading int")
)

// P(n)Y(n)M(n)DT(n)H(n)M(n)S.
var (
	defaultOption = duration{
		from: time.Now,
	}
	units = map[string]func(from time.Time, v uint64, scale float64) uint64{
		"M": month,
		"Y": year,
	}
	dateUnit = sampleUnits{
		"D": uint64(time.Hour * 24),
	}
	timeUnits = sampleUnits{
		"S": uint64(time.Second),
		"M": uint64(time.Minute),
		"H": uint64(time.Hour),
	}
)

func From(from func() time.Time) Option {
	return func(d *duration) {
		d.from = from
	}
}

type Option func(*duration)

// ParseDuration parses a duration string format P(n)Y(n)M(n)DT(n)H(n)M(n)S.
// use iso8601.From(time) when using the month and year, by default time.Now().
func ParseDuration(s string, opts ...Option) (time.Duration, error) {
	option := defaultOption
	for _, opt := range opts {
		opt(&option)
	}

	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	if s == "" {
		return 0, fmt.Errorf("iso8601: empty %w %q", ErrInvalidDuration, orig)
	}

	if s[0] != 'P' {
		return 0, fmt.Errorf("iso8601: format %w %q", ErrInvalidDuration, orig)
	}

	s = s[1:]
	unit := option.unit

	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		if s != "" && s[0] == 'T' {
			s = s[1:]
			unit = timeUnits.unit
		}

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, fmt.Errorf("iso8601: next character %w %q", ErrInvalidDuration, orig)
		}

		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, fmt.Errorf("iso8601: leadingInt %w %q", ErrInvalidDuration, orig)
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}

		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, fmt.Errorf("iso8601: leadingFraction %w %q", ErrInvalidDuration, orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' || c == 'T' {
				break
			}
		}
		if i == 0 {
			return 0, fmt.Errorf("iso8601: %w %q", ErrMissingUnit, orig)
		}
		u := s[:i]
		s = s[i:]

		v, err = unit(u, v, 0)
		if err != nil {
			return 0, fmt.Errorf("iso8601: %w unit %q", err, orig)
		}

		if f > 0 {
			r, err := unit(u, f, scale)
			if err != nil {
				return 0, fmt.Errorf("iso8601: %w fraction %q", err, orig)
			}
			log.Println(u, f, scale, r)

			v += r
		}

		if d > 1<<63-v {
			return 0, fmt.Errorf("iso8601: 1<<63 %w %q", ErrOverflow, orig)
		}
		d += v
	}

	if neg {
		return -time.Duration(d), nil
	}

	if d > 1<<63-1 {
		return 0, fmt.Errorf("iso8601: %w %q", ErrOverflow, orig)
	}

	return time.Duration(d), nil
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x uint64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, "", ErrLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, "", ErrLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

func month(from time.Time, v uint64, scale float64) uint64 {
	if scale == 0 {
		return uint64(from.AddDate(0, int(v), 0).Sub(from))
	}

	return uint64(float64(v) * (float64(from.AddDate(0, 1, 0).Sub(from)) / scale))
}

func year(from time.Time, v uint64, scale float64) uint64 {
	if scale == 0 {
		return uint64(from.AddDate(int(v), 0, 0).Sub(from))
	}

	return uint64(float64(v) * (float64(from.AddDate(1, 0, 0).Sub(from)) / scale))
}

type sampleUnits map[string]uint64

func (s sampleUnits) unit(name string, v uint64, scale float64) (uint64, error) {
	if unit, ok := s[name]; ok {
		if scale != 0 {
			v = uint64(float64(v) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, fmt.Errorf("iso8601:%w", ErrOverflow)
			}

			return v, nil
		}

		if v > 1<<63/unit {
			// overflow
			return 0, fmt.Errorf("iso8601:%w", ErrOverflow)
		}

		return v * unit, nil
	}

	return 0, fmt.Errorf("iso8601:%w", ErrMissingUnit)
}

type duration struct {
	from func() time.Time
}

func (d *duration) unit(name string, v uint64, scale float64) (uint64, error) {
	if _, ok := dateUnit[name]; ok {
		return dateUnit.unit(name, v, scale)
	}

	if unit, ok := units[name]; ok {
		from := d.from()
		out := unit(from, v, scale)
		if out > 1<<63 {
			// overflow
			return 0, ErrOverflow
		}

		d.from = func() time.Time {
			return from.Add(time.Duration(out))
		}

		return out, nil
	}

	return 0, fmt.Errorf("%w %q", ErrMissingUnit, name)
}

// FormatDuration returns a string representing the duration in the form "P1Y2M3DT4H5M6S".
// Leading zero units are omitted. The zero duration formats as PT0S.
func FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return "PT0S"
	}

	var buf [32]byte
	w := len(buf)
	u := uint64(duration)
	neg := duration < 0
	if neg {
		u = -u
	}

	w--
	buf[w] = 'S'
	w, u = fmtFrac(buf[:w], u, 9)

	// u is now integer seconds
	w = fmtInt(buf[:w], u%60)

	if u%60 == 0 && w+2 == len(buf) {
		w += 2
	}
	u /= 60

	// u is now integer minutes
	if u > 0 {
		if u%60 > 0 {
			w--
			buf[w] = 'M'
			w = fmtInt(buf[:w], u%60)
		}
		u /= 60

		if u > 0 && u%24 > 0 {
			w--
			buf[w] = 'H'
			w = fmtInt(buf[:w], u%24)
		}
		u /= 24
	}

	if w != len(buf) {
		w--
		buf[w] = 'T'
	}

	if u > 0 {
		w--
		buf[w] = 'D'
		w = fmtInt(buf[:w], u)
	}

	w--
	buf[w] = 'P'

	if neg {
		w--
		buf[w] = '-'
	}

	return string(buf[w:])
}

func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	isPrint := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		isPrint = isPrint || digit != 0
		if isPrint {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if isPrint {
		w--
		buf[w] = '.'
	}
	return w, v
}
