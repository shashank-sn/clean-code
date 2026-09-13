package adaptive

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateCalendarDate accepts only real Gregorian YYYY-MM-DD values.
// It rejects impossible dates such as 2026-02-30 instead of normalizing them.
func ValidateCalendarDate(value string) error {
	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return fmt.Errorf("calendar date must be YYYY-MM-DD, got %q", value)
	}
	if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return fmt.Errorf("calendar date must be YYYY-MM-DD with zero-padded month and day, got %q", value)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1 {
		return fmt.Errorf("calendar year is invalid in %q", value)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return fmt.Errorf("calendar month is invalid in %q", value)
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 {
		return fmt.Errorf("calendar day is invalid in %q", value)
	}
	maxDay := daysInMonth(year, month)
	if day > maxDay {
		return fmt.Errorf("calendar date %q does not exist", value)
	}
	return nil
}

func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}
