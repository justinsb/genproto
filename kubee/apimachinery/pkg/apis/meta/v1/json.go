package v1

import (
	"encoding/json"
	"time"
)

// UnmarshalJSON implements the json.Unmarshaller interface.
func (t *Time) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		*t = Time{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}

	seconds := pt.Unix()
	nanos := int32(pt.Nanosecond())
	t.Seconds = &seconds
	t.Nanos = &nanos
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.GetNanos() == 0 && t.GetSeconds() == 0 {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}
	buf := make([]byte, 0, len(time.RFC3339)+2)
	buf = append(buf, '"')

	tt := time.Unix(t.GetSeconds(), int64(t.GetNanos()))

	// time cannot contain non escapable JSON characters
	buf = tt.UTC().AppendFormat(buf, time.RFC3339)
	buf = append(buf, '"')
	return buf, nil
}
