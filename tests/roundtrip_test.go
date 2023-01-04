package tests

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	kubeeeappsv1 "justinsb.com/kubee/api/apps/v1"
	kubeeecorev1 "justinsb.com/kubee/api/core/v1"
	kubeeruntime "justinsb.com/kubee/apimachinery/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/davecgh/go-spew/spew"
	legacyproto "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	genericfuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	kubeproto "k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/apimachinery/pkg/util/diff"
)

var groups = []runtime.SchemeBuilder{
	appsv1.SchemeBuilder,
	corev1.SchemeBuilder,
}

func TestRoundTrip(t *testing.T) {
	h := NewFuzzHarness(t)

	seed := rand.Int63()
	fuzzer := fuzzer.FuzzerFor(genericfuzzer.Funcs, rand.NewSource(seed), h.codecs)

	for gvk := range h.scheme.AllKnownTypes() {
		if gvk.Version == "__internal" {
			continue
		}
		name := gvk.Group
		if name == "" {
			name = "core"
		}
		name += "." + gvk.Version + "." + gvk.Kind
		if gvk.Kind == "WatchEvent" {
			// Does not round trip?
			continue
		}
		for encoding := range h.encodings {
			h.Run(encoding+"/"+name, func(h *FuzzHarness) {
				var object runtime.Object
				object, err := h.scheme.New(gvk)
				if err != nil {
					t.Fatalf("Couldn't make a %v: %v", gvk, err)
				}

				fuzzer.Fuzz(object)

				accessor, err := apimeta.TypeAccessor(object)
				if err != nil {
					t.Fatalf("accessor failed: %v", err)
				}
				apiVersion, kind := gvk.ToAPIVersionAndKind()
				accessor.SetAPIVersion(apiVersion)
				accessor.SetKind(kind)

				switch encoding {
				case "json":
					h.TestRoundTripJSON(object)
				case "proto":
					h.TestRoundTripProto(object)
				}

			})
		}
	}
}

type FuzzHarness struct {
	*testing.T

	scheme    *runtime.Scheme
	codecs    serializer.CodecFactory
	encodings map[string]runtime.Codec
}

func (h *FuzzHarness) Run(name string, fn func(h *FuzzHarness)) {
	h.T.Run(name, func(t *testing.T) {
		h2 := *h
		h2.T = t
		fn(&h2)
	})
}
func NewFuzzHarness(t *testing.T) *FuzzHarness {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	for _, builder := range groups {
		require.NoError(t, builder.AddToScheme(scheme))
	}

	encodings := map[string]runtime.Codec{
		"json":  kubejson.NewSerializer(kubejson.DefaultMetaFactory, scheme, scheme, false),
		"proto": kubeproto.NewSerializer(scheme, scheme),
	}

	return &FuzzHarness{
		T:         t,
		codecs:    codecs,
		scheme:    scheme,
		encodings: encodings,
	}
}
func (h *FuzzHarness) TestRoundTripJSON(object runtime.Object) {
	codec := h.encodings["json"]
	h.testRoundTrip(codec, object, json.Unmarshal, json.Marshal)
}

func (h *FuzzHarness) TestRoundTripProto(object runtime.Object) {
	codec := h.encodings["proto"]
	protoMarshal := func(obj interface{}) ([]byte, error) {
		envelope := &kubeeruntime.Unknown{}

		raw, err := proto.Marshal(obj.(proto.Message))
		if err != nil {
			return nil, fmt.Errorf("error marshaling %T: %w", obj, err)
		}
		envelope.Raw = raw

		accessor, err := apimeta.TypeAccessor(object)
		if err != nil {
			return nil, fmt.Errorf("apimeta.TypeAccessor failed: %w", err)
		}
		apiVersion := accessor.GetAPIVersion()
		kind := accessor.GetKind()
		envelope.TypeMeta = &kubeeruntime.TypeMeta{
			ApiVersion: &apiVersion,
			Kind:       &kind,
		}
		b, err := proto.Marshal(envelope)
		if err != nil {
			return nil, fmt.Errorf("error marshaling envelope: %w", err)
		}

		magic := []byte{0x6b, 0x38, 0x73, 0x00}
		b2 := make([]byte, len(b)+4)
		copy(b2, magic)
		copy(b2[4:], b)
		return b2, nil
	}
	protoUnmarshal := func(data []byte, obj interface{}) error {
		if len(data) < 4 {
			return fmt.Errorf("data too short")
		}
		if data[0] != 0x6b || data[1] != 0x38 || data[2] != 0x73 || data[3] != 0 {
			return fmt.Errorf("corrupt proto data (bad magic)")
		}
		data = data[4:]

		envelope := &kubeeruntime.Unknown{}

		if err := proto.Unmarshal(data, envelope); err != nil {
			return fmt.Errorf("error unmarshaling to runtime.Unknown: %w", err)
		}

		if err := proto.Unmarshal(envelope.GetRaw(), obj.(proto.Message)); err != nil {
			return fmt.Errorf("error unmarshaling to %T: %w", obj, err)
		}
		return nil
	}
	h.testRoundTrip(codec, object, protoUnmarshal, protoMarshal)
}

func (t *FuzzHarness) encodeWithCodec(codec runtime.Codec, object runtime.Object) []byte {
	printer := spew.ConfigState{DisableMethods: true}

	original := object

	// deep copy the original object
	object = object.DeepCopyObject()
	name := reflect.TypeOf(object).Elem().Name()
	if !apiequality.Semantic.DeepEqual(original, object) {
		t.Errorf("%v: DeepCopy altered the object, diff: %v", name, diff.ObjectReflectDiff(original, object))
		t.Errorf("%s", spew.Sdump(original))
		t.Errorf("%s", spew.Sdump(object))
	}

	// encode (serialize) the deep copy using the provided codec
	data, err := runtime.Encode(codec, object)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			t.Fatalf("%v: not registered: %v (%s)", name, err, printer.Sprintf("%#v", object))
		} else {
			t.Errorf("%v: %v (%s)", name, err, printer.Sprintf("%#v", object))
		}
	}

	// ensure that the deep copy is equal to the original; neither the deep
	// copy or conversion should alter the object
	// TODO eliminate this global
	if !apiequality.Semantic.DeepEqual(original, object) {
		t.Fatalf("%v: encode altered the object, diff: %v", name, diff.ObjectReflectDiff(original, object))
	}

	// encode (serialize) a second time to verify that it was not varying
	secondData, err := runtime.Encode(codec, object)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			t.Fatalf("%v: not registered: %v (%s)", name, err, printer.Sprintf("%#v", object))
		} else {
			t.Fatalf("%v: %v (%s)", name, err, printer.Sprintf("%#v", object))
		}
	}

	// serialization to the wire must be stable to ensure that we don't write twice to the DB
	// when the object hasn't changed.
	if !bytes.Equal(data, secondData) {
		t.Fatalf("%v: serialization is not stable: %s", name, printer.Sprintf("%#v", object))
	}

	// decode (deserialize) the encoded data back into an object
	obj2, err := runtime.Decode(codec, data)
	if err != nil {
		t.Fatalf("FAIL:ERROR: %v: %v\nCodec: %#v\nData: %s\nSource: %#v", name, err, codec, dataAsString(data), printer.Sprintf("%#v", object))
	}

	// ensure that the object produced from decoding the encoded data is equal
	// to the original object
	if !apiequality.Semantic.DeepEqual(original, obj2) {
		t.Fatalf("FAIL:DIFF: %v: diff: %v\nCodec: %#v\nSource:\n\n%#v\n\nEncoded:\n\n%s\n\nFinal:\n\n%#v", name, diff.ObjectReflectDiff(original, obj2), codec, printer.Sprintf("%#v", original), dataAsString(data), printer.Sprintf("%#v", obj2))
	}

	return data
}

func (t *FuzzHarness) testRoundTrip(codec runtime.Codec, object runtime.Object, unmarshal func([]byte, interface{}) error, marshal func(obj interface{}) ([]byte, error)) {
	printer := spew.ConfigState{DisableMethods: true}
	original := object

	data := t.encodeWithCodec(codec, object)

	object = object.DeepCopyObject()

	var kubee interface{}
	gvk := object.GetObjectKind().GroupVersionKind()
	switch gvk.Kind {
	case "ConfigMap":
		kubee = &kubeeecorev1.ConfigMap{}
	case "Pod":
		kubee = &kubeeecorev1.Pod{}
	case "Secret":
		kubee = &kubeeecorev1.Secret{}
	case "Deployment":
		kubee = &kubeeeappsv1.Deployment{}
	}

	if kubee != nil {
		t.Logf("kubee input: %v", string(data))
		if err := unmarshal(data, kubee); err != nil {
			t.Fatalf("FAIL:ERROR: error unmarshaling: %v", err)
		}

		backToBytes, err := marshal(kubee)
		if err != nil {
			t.Fatalf("error marshaling: %v", err)
		}
		t.Logf("kubee output: %v", string(backToBytes))

		// decode (deserialize) the encoded data back into an object
		obj3, err := runtime.Decode(codec, backToBytes)
		if err != nil {
			t.Fatalf("FAIL:ERROR: %v\nCodec: %#v\nData: %s\nSource: %#v", err, codec, dataAsString(data), printer.Sprintf("%#v", object))
		}

		// ensure that the object produced from decoding the encoded data is equal
		// to the original object
		if !apiequality.Semantic.DeepEqual(original, obj3) {
			t.Fatalf("FAIL:DIFF: %v\nCodec: %#v\nSource:\n\n%#v\n\nEncoded:\n\n%s\n\nFinal:\n\n%#v", diff.ObjectReflectDiff(original, obj3), codec, printer.Sprintf("%#v", original), dataAsString(data), printer.Sprintf("%#v", obj3))
		}
	}
}

// dataAsString returns the given byte array as a string; handles detecting
// protocol buffers.
func dataAsString(data []byte) string {
	dataString := string(data)
	if !strings.HasPrefix(dataString, "{") {
		dataString = "\n" + hex.Dump(data)
		legacyproto.NewBuffer(make([]byte, 0, 1024)).DebugPrint("decoded object", data)
	}
	return dataString
}
