package audit

import "testing"

// A file written into an image or deleted from one changes what the host will
// see, and so does a new exchange image: each is a media event with a name of
// its own. Browsing and downloading are reads and leave no line.
func TestImageFileChangesAreMediaEvents(t *testing.T) {
	for route, action := range map[string]string{
		"/api/storage/files/upload": "file-upload",
		"/api/storage/files/delete": "file-delete",
		"/api/storage/image/create": "create-image",
	} {
		rule, ok := Classify("POST", route)
		if !ok || rule.Category != CategoryMedia || rule.Action != action {
			t.Errorf("POST %s: %+v, %v; want %s/%s", route, rule, ok, CategoryMedia, action)
		}
	}
}
