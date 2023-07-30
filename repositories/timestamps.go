package repositories

import "time"

type Timestamps struct {
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Deleted time.Time `json:"deleted"`
}

func Now() *Timestamps {
	return &Timestamps{
		Created: time.Now(),
		Updated: time.Now(),
	}
}

func (t *Timestamps) Update() {
	t.Updated = time.Now()
}

// For soft deletes if the future
func (t *Timestamps) Delete() {
	t.Deleted = time.Now()
}
