package runtime

import (
	"bytes"
	"errors"
)

func (re *RawExtension) UnmarshalJSON(in []byte) error {
	if re == nil {
		return errors.New("runtime.RawExtension: UnmarshalJSON on nil pointer")
	}
	if !bytes.Equal(in, []byte("null")) {
		re.Raw = append(re.Raw[0:0], in...)
	}
	return nil
}

// MarshalJSON may get called on pointers or values, so implement MarshalJSON on value.
// http://stackoverflow.com/questions/21390979/custom-marshaljson-never-gets-called-in-go
func (re RawExtension) MarshalJSON() ([]byte, error) {
	if re.Raw == nil {
		return []byte("null"), nil
	}
	// TODO: Check whether ContentType is actually JSON before returning it.
	return re.Raw, nil
}
