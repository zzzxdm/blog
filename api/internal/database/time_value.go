package database

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FlexibleTime struct {
	time.Time
}

func (value *FlexibleTime) Scan(input any) error {
	switch typed := input.(type) {
	case time.Time:
		value.Time = typed
		return nil
	case string:
		return value.scanString(typed)
	case []byte:
		return value.scanString(string(typed))
	case int64:
		value.Time = time.Unix(typed, 0)
		return nil
	case nil:
		value.Time = time.Time{}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into FlexibleTime", input)
	}
}

func (value FlexibleTime) Value() (driver.Value, error) {
	if value.Time.IsZero() {
		return nil, nil
	}
	return value.Time, nil
}

func (value *FlexibleTime) scanString(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		value.Time = time.Time{}
		return nil
	}
	if unix, err := strconv.ParseInt(input, 10, 64); err == nil {
		value.Time = time.Unix(unix, 0)
		return nil
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999999",
		time.DateTime,
		time.DateOnly,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, input); err == nil {
			value.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse time %q", input)
}
