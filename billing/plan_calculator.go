package billing

import (
	"math"
)

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func PlanAndDurationToPrice(plan string, duration string) float64 {
	price := float64(0.00)
	if duration == "monthly" {
		switch plan {
		case "Personal": // now "Personal"
			price = 18.99 * 1
		case "Consultant": // now "Consultant"
			price = 34.99 * 1
		case "Business": // now "Business"
			price = 41.99 * 1
		case "Growing Business": // now "Growing Business"
			price = 52.99 * 1
		}
	} else {
		switch plan {
		case "Personal": // now "Personal"
			price = 15.99 * 12
		case "Consultant": // now "Consultant"
			price = 28.99 * 12
		case "Business": // now "Business"
			price = 34.99 * 12
		case "Growing Business": // now "Growing Business"
			price = 43.99 * 12
		}
	}

	return toFixed(price, 2)
}
