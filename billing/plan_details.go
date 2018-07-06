package billing

func BillingIdToPlanName(plan string) string {
	switch plan {
	case "bronze", "personal": // now "Personal"
		return "Personal"
	case "aluminum", "consultant": // now "Consultant"
		return "Consultant"
	case "silver", "silver-1", "business": // now "Business"
		return "Business"
	case "gold", "gold-1", "growing": // now "Growing Business"
		return "Growing Business"
	case "free":
		return "Trial Member"
	}

	return "Personal"
}

func UserMaximumSocialAccounts(plan string) int {
	switch plan {
	case "Personal": // now "Personal"
		return 100
	case "Consultant": // now "Consultant"
		return 250
	case "Business": // now "Business"
		return 500
	case "Growing Business": // now "Growing Business"
		return 100000
	}

	return 0
}

func UserMaximumEmailAccounts(plan string) int {
	switch plan {
	case "Personal": // now "Personal"
		return 0
	case "Consultant": // now "Consultant"
		return 2
	case "Business": // now "Business"
		return 5
	case "Growing Business": // now "Growing Business"
		return 10
	}

	return 0
}

func UserMaximumEmailSent(plan string) int {
	switch plan {
	case "Personal": // now "Personal"
		return 100
	case "Consultant": // now "Consultant"
		return 400
	case "Business": // now "Business"
		return 1000
	case "Growing Business": // now "Growing Business"
		return 2500
	}

	return 0
}

func StripePlanIdToMaximumEmailSent(stripePlanId string) int {
	switch stripePlanId {
	case "free": // now "Personal"
		return 100
	}

	return 20000
}
