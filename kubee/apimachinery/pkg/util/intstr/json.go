package intstr

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Type represents the stored type of IntOrString.
type Type int64

const (
	Int    Type = iota // The IntOrString holds an int.
	String             // The IntOrString holds a string.
)

// UnmarshalJSON implements the json.Unmarshaller interface.
func (intstr *IntOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		t := int64(String)
		intstr.Type = &t
		return json.Unmarshal(value, &intstr.StrVal)
	}
	t := int64(Int)
	intstr.Type = &t
	return json.Unmarshal(value, &intstr.IntVal)
}

// IntValue returns the IntVal if type Int, or if
// it is a String, will attempt a conversion to int,
// returning 0 if a parsing error occurs.
func (intstr *IntOrString) IntValue() int {
	if Type(intstr.GetType()) == String {
		i, _ := strconv.Atoi(intstr.GetStrVal())
		return i
	}
	return int(intstr.GetIntVal())
}

// MarshalJSON implements the json.Marshaller interface.
func (intstr IntOrString) MarshalJSON() ([]byte, error) {
	switch Type(intstr.GetType()) {
	case Int:
		return json.Marshal(intstr.GetIntVal())
	case String:
		return json.Marshal(intstr.GetStrVal())
	default:
		return []byte{}, fmt.Errorf("impossible IntOrString.Type")
	}
}
