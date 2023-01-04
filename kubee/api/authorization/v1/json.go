package v1

import (
	"encoding/json"
)

// UnmarshalJSON implements the json.Unmarshaller interface.
func (v *ExtraValue) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		*v = ExtraValue{}
		return nil
	}

	var items []string
	err := json.Unmarshal(b, &items)
	if err != nil {
		return err
	}

	v.Items = items

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (v ExtraValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Items)
}
