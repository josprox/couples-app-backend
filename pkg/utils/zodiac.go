package utils

import "time"

func CalculateZodiac(birthDate time.Time) string {
	day := birthDate.Day()
	month := birthDate.Month()

	switch month {
	case time.January:
		if day <= 19 {
			return "Capricorn"
		}
		return "Aquarius"
	case time.February:
		if day <= 18 {
			return "Aquarius"
		}
		return "Pisces"
	case time.March:
		if day <= 20 {
			return "Pisces"
		}
		return "Aries"
	case time.April:
		if day <= 19 {
			return "Aries"
		}
		return "Taurus"
	case time.May:
		if day <= 20 {
			return "Taurus"
		}
		return "Gemini"
	case time.June:
		if day <= 20 {
			return "Gemini"
		}
		return "Cancer"
	case time.July:
		if day <= 22 {
			return "Cancer"
		}
		return "Leo"
	case time.August:
		if day <= 22 {
			return "Leo"
		}
		return "Virgo"
	case time.September:
		if day <= 22 {
			return "Virgo"
		}
		return "Libra"
	case time.October:
		if day <= 22 {
			return "Libra"
		}
		return "Scorpio"
	case time.November:
		if day <= 21 {
			return "Scorpio"
		}
		return "Sagittarius"
	case time.December:
		if day <= 21 {
			return "Sagittarius"
		}
		return "Capricorn"
	default:
		return "Unknown"
	}
}
