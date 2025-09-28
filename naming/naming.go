package naming

import (
	"strings"
)

const IdSeparator = "-"

func GetId(profile, vendor, name string) string {
	profile = SanitizeProfile(profile)
	vendor = Sanitize(vendor)
	name = Sanitize(name)
	return strings.ToLower(strings.Join([]string{profile, vendor, name}, IdSeparator))
}
