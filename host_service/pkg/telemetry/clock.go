package telemetry

import (
	"strings"
	"time"

	"monitor-esp32-host/pkg/models"
)

var spanishDays = map[string]string{
	"Mon": "LUN", "Tue": "MAR", "Wed": "MIE", "Thu": "JUE",
	"Fri": "VIE", "Sat": "SAB", "Sun": "DOM",
}

var spanishMonths = map[string]string{
	"Jan": "ENE", "Feb": "FEB", "Mar": "MAR", "Apr": "ABR",
	"May": "MAY", "Jun": "JUN", "Jul": "JUL", "Aug": "AGO",
	"Sep": "SEP", "Oct": "OCT", "Nov": "NOV", "Dec": "DIC",
}

// GetClockMetrics devuelve la hora y fecha actual de la computadora formateada.
func GetClockMetrics() models.ClockMetrics {
	now := time.Now()

	dayEng := now.Format("Mon")
	daySpa := spanishDays[dayEng]
	if daySpa == "" {
		daySpa = strings.ToUpper(dayEng)
	}

	monthEng := now.Format("Jan")
	monthSpa := spanishMonths[monthEng]
	if monthSpa == "" {
		monthSpa = strings.ToUpper(monthEng)
	}

	return models.ClockMetrics{
		Time: now.Format("15:04"),
		Date: now.Format("02 ") + monthSpa,
		Day:  daySpa,
	}
}

