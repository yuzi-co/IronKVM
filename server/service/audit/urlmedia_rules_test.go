package audit

import "testing"

// Mounting an image from a URL and ejecting it change what the host sees,
// so both are media events with names of their own (ironkvm-dist#106).
func TestURLMediaChangesAreMediaEvents(t *testing.T) {
	for route, action := range map[string]string{
		"/api/storage/url/mount": "mount-url",
		"/api/storage/url/eject": "eject-url",
	} {
		rule, ok := Classify("POST", route)
		if !ok || rule.Category != CategoryMedia || rule.Action != action {
			t.Errorf("POST %s: %+v, %v; want %s/%s", route, rule, ok, CategoryMedia, action)
		}
	}
}
