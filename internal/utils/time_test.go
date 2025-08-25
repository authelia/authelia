package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDurationString(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		raw      bool
		expected time.Duration
		err      string
	}{
		{"ShouldParseStringsForMillisecond", []string{"%d ms", "%d millisecond", "%d milliseconds"}, false, time.Millisecond, ""},
		{"ShouldParseStringsForSecond", []string{"%d s", "%d second", "%d seconds"}, false, time.Second, ""},
		{"ShouldParseStringsForMinute", []string{"%d m", "%d minute", "%d minutes"}, false, time.Minute, ""},
		{"ShouldParseStringsForHour", []string{"%d h", "%d hour", "%d hours"}, false, time.Hour, ""},
		{"ShouldParseStringsForDay", []string{"%d d", "%d day", "%d days"}, false, time.Hour * HoursInDay, ""},
		{"ShouldParseStringsForWeek", []string{"%d w", "%d week", "%d weeks"}, false, time.Hour * HoursInWeek, ""},
		{"ShouldParseStringsForMonth", []string{"%d M", "%d month", "%d months"}, false, time.Hour * HoursInMonth, ""},
		{"ShouldParseStringsForYear", []string{"%d y", "%d year", "%d years"}, false, time.Hour * HoursInYear, ""},
		{"ShouldParseStringsDecimals", []string{"100"}, true, time.Second * 100, ""},
		{"ShouldParseStringsDecimalNull", []string{""}, true, time.Second * 0, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, f := range tc.have {
				if tc.raw {
					t.Run(f, func(t *testing.T) {
						actual, actualErr := ParseDurationString(f)

						if tc.err == "" {
							assert.NoError(t, actualErr)
							assert.Equal(t, tc.expected, actual)
						} else {
							assert.EqualError(t, actualErr, tc.err)
						}
					})
				} else {
					for _, d := range []int{1, 5, 20} {
						input := fmt.Sprintf(f, d)

						inputNoSpace := strings.ReplaceAll(input, " ", "")

						t.Run(inputNoSpace, func(t *testing.T) {
							t.Run("WithSpaces", func(t *testing.T) {
								actual, actualErr := ParseDurationString(input)

								if tc.err == "" {
									assert.NoError(t, actualErr)
									assert.Equal(t, tc.expected*time.Duration(d), actual)
								} else {
									assert.EqualError(t, actualErr, tc.err)
								}

								t.Run("LeadingZeros", func(t *testing.T) {
									inputActual := reNumeric.ReplaceAllStringFunc(input, func(s string) string {
										return "000" + s
									})

									actual, actualErr := ParseDurationString(inputActual)

									if tc.err == "" {
										assert.NoError(t, actualErr)
										assert.Equal(t, tc.expected*time.Duration(d), actual)
									} else {
										assert.EqualError(t, actualErr, tc.err)
									}
								})
							})

							t.Run("WithoutSpaces", func(t *testing.T) {
								actual, actualErr := ParseDurationString(inputNoSpace)

								if tc.err == "" {
									assert.NoError(t, actualErr)
									assert.Equal(t, tc.expected*time.Duration(d), actual)
								} else {
									assert.EqualError(t, actualErr, tc.err)
								}

								t.Run("LeadingZeros", func(t *testing.T) {
									inputActual := reNumeric.ReplaceAllStringFunc(inputNoSpace, func(s string) string {
										return "000" + s
									})

									actual, actualErr := ParseDurationString(inputActual)

									if tc.err == "" {
										assert.NoError(t, actualErr)
										assert.Equal(t, tc.expected*time.Duration(d), actual)
									} else {
										assert.EqualError(t, actualErr, tc.err)
									}
								})
							})
						})
					}
				}
			}
		})
	}
}

func TestStandardizeDurationString(t *testing.T) {
	var (
		actual string
		err    error
	)

	actual, err = StandardizeDurationString("1 hour and 20 minutes")

	assert.NoError(t, err)
	assert.Equal(t, "1h20m", actual)

	actual, err = StandardizeDurationString("1 hour    and 20 minutes")

	assert.NoError(t, err)
	assert.Equal(t, "1h20m", actual)
}

func TestParseDurationString_ShouldNotParseDurationStringWithOutOfOrderQuantitiesAndUnits(t *testing.T) {
	duration, err := ParseDurationString("h1")

	assert.EqualError(t, err, "could not parse 'h1' as a duration")
	assert.Equal(t, time.Duration(0), duration)
}

func TestParseDurationString_ShouldNotParseBadDurationString(t *testing.T) {
	duration, err := ParseDurationString("10x")

	assert.EqualError(t, err, "could not parse the units portion of '10x' in duration string '10x': the unit 'x' is not valid")
	assert.Equal(t, time.Duration(0), duration)
}

func TestParseDurationString_ShouldNotParseBadDurationStringAlt(t *testing.T) {
	duration, err := ParseDurationString("10abcxyz")

	assert.EqualError(t, err, "could not parse the units portion of '10abcxyz' in duration string '10abcxyz': the unit 'abcxyz' is not valid")
	assert.Equal(t, time.Duration(0), duration)
}

func TestParseDurationString_ShouldParseMultiUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d3w10ms")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(7)*time.Duration(3))+
			(time.Millisecond*time.Duration(10)), duration)
}

func TestParseDurationString_ShouldParseDuplicateUnitValues(t *testing.T) {
	duration, err := ParseDurationString("1d4d2d")

	assert.NoError(t, err)
	assert.Equal(t,
		(time.Hour*time.Duration(24))+
			(time.Hour*time.Duration(24)*time.Duration(4))+
			(time.Hour*time.Duration(24)*time.Duration(2)), duration)
}

func TestStandardizeDurationString_ShouldParseStringWithSpaces(t *testing.T) {
	result, err := StandardizeDurationString("1d 1h 20m")

	assert.NoError(t, err)
	assert.Equal(t, result, "24h1h20m")
}

func TestShouldTimeIntervalsMakeSense(t *testing.T) {
	assert.Equal(t, Hour, time.Minute*60)
	assert.Equal(t, Day, Hour*24)
	assert.Equal(t, Week, Day*7)
	assert.Equal(t, Year, Day*365)
	assert.Equal(t, Month, Year/12)
}

func TestShouldConvertKnownUnixNanoTimeToKnownWin32Epoch(t *testing.T) {
	exampleNanoTime := int64(1626234411 * 1000000000)
	win32Epoch := uint64(132707080110000000)

	assert.Equal(t, win32Epoch, UnixNanoTimeToMicrosoftNTEpoch(exampleNanoTime))
	assert.Equal(t, timeUnixEpochAsMicrosoftNTEpoch, UnixNanoTimeToMicrosoftNTEpoch(0))
	assert.Equal(t, timeUnixEpochAsMicrosoftNTEpoch, UnixNanoTimeToMicrosoftNTEpoch(-1))
}

func TestParseTimeString(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		index    int
		expected time.Time
		err      string
	}{
		{"ShouldParseIntegerAsUnix", "1675899060", -1, time.Unix(1675899060, 0), ""},
		{"ShouldParseIntegerAsUnixMilli", "1675899060000", -2, time.Unix(1675899060, 0), ""},
		{"ShouldParseIntegerAsUnixMicro", "1675899060000000", -3, time.Unix(1675899060, 0), ""},
		{"ShouldNotParseSuperLargeInteger", "9999999999999999999999999999999999999999", -999, time.Unix(0, 0), "time value was detected as an integer but the integer could not be parsed: strconv.ParseInt: parsing \"9999999999999999999999999999999999999999\": value out of range"},
		{"ShouldParseSimpleTime", "Jan 2 15:04:05 2006", 0, time.Unix(1136214245, 0), ""},
		{"ShouldNotParseInvalidTime", "abc", -998, time.Unix(0, 0), "failed to find a suitable time layout for time 'abc'"},
		{"ShouldMatchDate", "2020-05-01", 6, time.Unix(1588291200, 0), ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			index, actualA, errA := matchParseTimeStringWithLayouts(tc.have, StandardTimeLayouts)
			actualB, errB := ParseTimeStringWithLayouts(tc.have, StandardTimeLayouts)
			actualC, errC := ParseTimeString(tc.have)

			if tc.err == "" {
				assert.NoError(t, errA)
				assert.NoError(t, errB)
				assert.NoError(t, errC)

				assert.Equal(t, tc.index, index)
				assert.Equal(t, tc.expected.UnixNano(), actualA.UnixNano())
				assert.Equal(t, tc.expected.UnixNano(), actualB.UnixNano())
				assert.Equal(t, tc.expected.UnixNano(), actualC.UnixNano())
			} else {
				assert.EqualError(t, errA, tc.err)
				assert.EqualError(t, errB, tc.err)
				assert.EqualError(t, errC, tc.err)
			}
		})
	}
}

func TestParseTimeStringWithLayouts(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		index    int
		expected time.Time
		err      string
	}{
		{"ShouldParseIntegerAsUnix", "1675899060", -1, time.Unix(1675899060, 0), ""},
		{"ShouldParseIntegerAsUnixMilli", "1675899060000", -2, time.Unix(1675899060, 0), ""},
		{"ShouldParseIntegerAsUnixMicro", "1675899060000000", -3, time.Unix(1675899060, 0), ""},
		{"ShouldNotParseSuperLargeInteger", "9999999999999999999999999999999999999999", -999, time.Unix(0, 0), "time value was detected as an integer but the integer could not be parsed: strconv.ParseInt: parsing \"9999999999999999999999999999999999999999\": value out of range"},
		{"ShouldParseSimpleTime", "Jan 2 15:04:05 2006", 0, time.Unix(1136214245, 0), ""},
		{"ShouldNotParseInvalidTime", "abc", -998, time.Unix(0, 0), "failed to find a suitable time layout for time 'abc'"},
		{"ShouldMatchDate", "2020-05-01", 6, time.Unix(1588291200, 0), ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseTimeStringWithLayouts(tc.have, StandardTimeLayouts)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.UnixNano(), actual.UnixNano())
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
