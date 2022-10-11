package iso8601_test

import (
	"fmt"
	"time"

	"gitoa.ru/go-4devs/iso8601"
)

func ExampleFormatDuration() {
	second := iso8601.FormatDuration(time.Second)
	hours := iso8601.FormatDuration(time.Hour * 25)
	partOfSecond := iso8601.FormatDuration(time.Second / 5)

	fmt.Printf("%s = %s\n", time.Second, second)
	fmt.Printf("%s = %s\n", time.Hour*25, hours)
	fmt.Printf("%s = %s\n", time.Second/5, partOfSecond)
	// Output:
	// 1s = PT1S
	// 25h0m0s = P1DT1H
	// 200ms = PT0.2S
}

func ExampleParseDuration() {
	year2020 := time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)
	second, errSecond := iso8601.ParseDuration("PT1S")
	hours, errHours := iso8601.ParseDuration("P1DT1H")
	partOfSecond, errPartOfSecond := iso8601.ParseDuration("PT0.2S")
	yearFromZeroTime, errYearFromZeroTime := iso8601.ParseDuration("P1Y1M1D", iso8601.From(func() time.Time { return time.Time{} }))
	yearFrom2020, errYearFrom2020 := iso8601.ParseDuration("P1Y1M1D", iso8601.From(func() time.Time { return year2020 }))

	fmt.Printf("PT1S = %s(%v)\n", second, errSecond)
	fmt.Printf("P1DT1H = %s(%v)\n", hours, errHours)
	fmt.Printf("PT0.2S = %s(%v)\n", partOfSecond, errPartOfSecond)
	fmt.Printf("P1Y1M1D(from %v) = %s(%v)\n", time.Time{}, yearFromZeroTime, errYearFromZeroTime)
	fmt.Printf("P1Y1M1D(form %v) = %s(%v)\n", year2020, yearFrom2020, errYearFrom2020)
	// Output:
	// PT1S = 1s(<nil>)
	// P1DT1H = 25h0m0s(<nil>)
	// PT0.2S = 200ms(<nil>)
	// P1Y1M1D(from 0001-01-01 00:00:00 +0000 UTC) = 9528h0m0s(<nil>)
	// P1Y1M1D(form 2020-01-01 01:01:01.000000001 +0000 UTC) = 9552h0m0s(<nil>)
}
