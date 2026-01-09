package duration

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(d).String()
	return json.Marshal(s)
}

type Timestamp struct {
	Minutes int `json:"minutes,omitempty"`
	Hours   int `json:"hours,omitempty"`
	Days    int `json:"days,omitempty"`
}

func (t *Timestamp) ToDuration() time.Duration {
	total := time.Duration(0)
	total += time.Duration(t.Minutes) * time.Minute
	total += time.Duration(t.Hours) * time.Hour
	total += time.Duration(t.Days) * 24 * time.Hour
	return total
}
