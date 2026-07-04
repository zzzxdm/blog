package operations

import (
	"fmt"
	"strings"
	"time"
)

type statsRangeSpec struct {
	Key   string
	Label string
	Start time.Time
	End   time.Time
}

func newStatsRangeSpec(rangeKey string, now time.Time) statsRangeSpec {
	key := normalizeStatsRange(rangeKey)
	end := now
	switch key {
	case "7d":
		return statsRangeSpec{
			Key:   key,
			Label: "最近 7 天",
			Start: now.AddDate(0, 0, -7),
			End:   end,
		}
	case "ytd":
		return statsRangeSpec{
			Key:   key,
			Label: "今年",
			Start: time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()),
			End:   end,
		}
	default:
		return statsRangeSpec{
			Key:   "30d",
			Label: "最近 30 天",
			Start: now.AddDate(0, 0, -30),
			End:   end,
		}
	}
}

func (spec statsRangeSpec) trendBucket() string {
	if spec.Key == "ytd" {
		return "month"
	}
	if spec.Key == "30d" {
		return "week"
	}
	return "day"
}

func formatTrendLabel(value time.Time, rangeKey string) string {
	if rangeKey == "ytd" {
		return fmt.Sprintf("%d月", int(value.Month()))
	}
	return value.Format("01-02")
}

func formatStatNumber(value int) string {
	if value == 0 {
		return "0"
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	digits := fmt.Sprintf("%d", value)
	first := len(digits) % 3
	if first == 0 {
		first = 3
	}

	var builder strings.Builder
	builder.WriteString(sign)
	builder.WriteString(digits[:first])
	for index := first; index < len(digits); index += 3 {
		builder.WriteString(",")
		builder.WriteString(digits[index : index+3])
	}

	return builder.String()
}

func formatReadingMinutes(value float64) string {
	if value <= 0 {
		return "0 分钟"
	}

	return fmt.Sprintf("%.1f 分钟", value)
}

func formatRate(numerator int, denominator int) string {
	if numerator <= 0 || denominator <= 0 {
		return "0%"
	}

	value := float64(numerator) * 100 / float64(denominator)
	if value >= 10 {
		return fmt.Sprintf("%.0f%%", value)
	}

	return fmt.Sprintf("%.1f%%", value)
}

func barPercent(value int, total int) int {
	if value <= 0 || total <= 0 {
		return 0
	}

	percent := (value*100 + total/2) / total
	if percent < 1 {
		return 1
	}
	if percent > 100 {
		return 100
	}

	return percent
}
