package jd_cookie

import (
	"time"
)

// This function counts the
// number of leap years
// since the starting of time
// to the current year that
// is passed
func leapYears(date time.Time) (leaps int) {

	// returns year, month,
	// date of a time object
	y, m, _ := date.Date()

	if m <= 2 {
		y--
	}
	leaps = y/4 + y/400 - y/100
	return leaps
}

// The function calculates the
// difference between two dates and times
// and returns the days, hours, minutes,
// seconds between two dates

func getDifference(a, b time.Time) (days, hours, minutes, seconds int) {

	// month-wise days
	monthDays := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// extracting years, months,
	// days of two dates
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()

	// extracting hours, minutes,
	// seconds of two times
	h1, min1, s1 := a.Clock()
	h2, min2, s2 := b.Clock()

	// totalDays since the
	// beginning = year*365 + number_of_days
	totalDays1 := y1*365 + d1

	// adding days of the months
	// before the current month
	for i := 0; i < (int)(m1)-1; i++ {
		totalDays1 += monthDays[i]
	}

	// counting leap years since
	// beginning to the year "a"
	// and adding that many extra
	// days to the totaldays
	totalDays1 += leapYears(a)

	// Similar procedure for second date
	totalDays2 := y2*365 + d2

	for i := 0; i < (int)(m2)-1; i++ {
		totalDays2 += monthDays[i]
	}

	totalDays2 += leapYears(b)

	// Number of days between two days
	days = totalDays2 - totalDays1

	// calculating hour, minutes,
	// seconds differences
	hours = h2 - h1
	minutes = min2 - min1
	seconds = s2 - s1

	// if seconds difference goes below 0,
	// add 60 and decrement number of minutes
	if seconds < 0 {
		seconds += 60
		minutes--
	}

	// performing similar operations
	// on minutes and hours
	if minutes < 0 {
		minutes += 60
		hours--
	}

	// performing similar operations
	// on hours and days
	if hours < 0 {
		hours += 24
		days--
	}

	return days, hours, minutes, seconds

}

// Driver code

// func main() {

// 	// Syntax for time date:
// 	// d := time.Date(year, month, days, hours,
// 	// minutes, seconds, nanoseconds, timeZone)

// 	date1 := time.Date(2020, 4, 27, 23, 35, 0, 0, time.UTC)
// 	date2 := time.Date(2018, 5, 12, 12, 43, 23, 0, time.UTC)

// 	// if date1 occurs after date2 then
// 	// swap days since absolute
// 	// difference is being calculated
// 	if date1.After(date2) {
// 		date1, date2 = date2, date1
// 	}
// 	// Calling function and getting
// 	// difference between two dates
// 	days, hours, minutes, seconds := getDifference(date1, date2)

// 	// Printing the difference
// 	fmt.Printf("%v days, %v hours, %v minutes, %v seconds", days, hours, minutes, seconds)

// }
