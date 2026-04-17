package notes

import "fmt"

func TodayTemplate(id string) string {
	return fmt.Sprintf("---\nid: %s\ntags: []\n---\n\n# %s\n\n", id, id)
}
