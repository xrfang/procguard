package procguard

type (
	slot struct {
		Since string
		Until string
	}
	slots map[string][]slot
)
