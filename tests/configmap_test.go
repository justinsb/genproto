package tests

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	kubeeecorev1 "justinsb.com/kubee/api/core/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	genericfuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/diff"
)

var groups = []runtime.SchemeBuilder{
	corev1.SchemeBuilder,
}

func TestConfigMap(t *testing.T) {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	for _, builder := range groups {
		require.NoError(t, builder.AddToScheme(scheme))
	}

	seed := rand.Int63()
	fuzzer := fuzzer.FuzzerFor(genericfuzzer.Funcs, rand.NewSource(seed), codecs)

	for gvk := range scheme.AllKnownTypes() {
		if gvk.Version == "__internal" {
			continue
		}
		name := gvk.Group
		if name == "" {
			name = "core"
		}
		name += "." + gvk.Version + "." + gvk.Kind
		if name == "core.v1.WatchEvent" {
			// Does not round trip?
			continue
		}
		t.Run(name, func(t *testing.T) {
			var object runtime.Object
			object, err := scheme.New(gvk)
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

			codec := kubejson.NewSerializer(kubejson.DefaultMetaFactory, scheme, scheme, false)

			testRoundTrip(t, codec, object)
		})
	}
}

func testRoundTrip(t *testing.T, codec runtime.Codec, object runtime.Object) {
	printer := spew.ConfigState{DisableMethods: true}

	original := object

	// deep copy the original object
	object = object.DeepCopyObject()
	name := reflect.TypeOf(object).Elem().Name()
	if !apiequality.Semantic.DeepEqual(original, object) {
		t.Errorf("%v: DeepCopy altered the object, diff: %v", name, diff.ObjectReflectDiff(original, object))
		t.Errorf("%s", spew.Sdump(original))
		t.Errorf("%s", spew.Sdump(object))
		return
	}

	// encode (serialize) the deep copy using the provided codec
	data, err := runtime.Encode(codec, object)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			t.Logf("%v: not registered: %v (%s)", name, err, printer.Sprintf("%#v", object))
		} else {
			t.Errorf("%v: %v (%s)", name, err, printer.Sprintf("%#v", object))
		}
		return
	}

	// ensure that the deep copy is equal to the original; neither the deep
	// copy or conversion should alter the object
	// TODO eliminate this global
	if !apiequality.Semantic.DeepEqual(original, object) {
		t.Errorf("%v: encode altered the object, diff: %v", name, diff.ObjectReflectDiff(original, object))
		return
	}

	// encode (serialize) a second time to verify that it was not varying
	secondData, err := runtime.Encode(codec, object)
	if err != nil {
		if runtime.IsNotRegisteredError(err) {
			t.Logf("%v: not registered: %v (%s)", name, err, printer.Sprintf("%#v", object))
		} else {
			t.Errorf("%v: %v (%s)", name, err, printer.Sprintf("%#v", object))
		}
		return
	}

	// serialization to the wire must be stable to ensure that we don't write twice to the DB
	// when the object hasn't changed.
	if !bytes.Equal(data, secondData) {
		t.Errorf("%v: serialization is not stable: %s", name, printer.Sprintf("%#v", object))
	}

	// decode (deserialize) the encoded data back into an object
	obj2, err := runtime.Decode(codec, data)
	if err != nil {
		t.Fatalf("%v: %v\nCodec: %#v\nData: %s\nSource: %#v", name, err, codec, dataAsString(data), printer.Sprintf("%#v", object))
	}

	// ensure that the object produced from decoding the encoded data is equal
	// to the original object
	if !apiequality.Semantic.DeepEqual(original, obj2) {
		t.Errorf("%v: diff: %v\nCodec: %#v\nSource:\n\n%#v\n\nEncoded:\n\n%s\n\nFinal:\n\n%#v", name, diff.ObjectReflectDiff(original, obj2), codec, printer.Sprintf("%#v", original), dataAsString(data), printer.Sprintf("%#v", obj2))
		return
	}

	var kubee interface{}
	gvk := object.GetObjectKind().GroupVersionKind()
	switch gvk.Kind {
	case "ConfigMap":
		kubee = &kubeeecorev1.ConfigMap{}
	case "Pod":
		kubee = &kubeeecorev1.Pod{}
	case "Secret":
		kubee = &kubeeecorev1.Secret{}
	}

	if kubee != nil {
		if err := json.Unmarshal(data, kubee); err != nil {
			t.Fatalf("error unmarshaling: %v", err)
		}
		t.Logf("kubee input json: %v", string(data))
		t.Logf("kubee from json: %+v", kubee)

		backToJSON, err := json.Marshal(kubee)
		if err != nil {
			t.Fatalf("error marshaling: %v", err)
		}
		t.Logf("kubee json output: %v", string(backToJSON))

		// decode (deserialize) the encoded data back into an object
		obj3, err := runtime.Decode(codec, backToJSON)
		if err != nil {
			t.Fatalf("%v: %v\nCodec: %#v\nData: %s\nSource: %#v", name, err, codec, dataAsString(data), printer.Sprintf("%#v", object))
		}

		// ensure that the object produced from decoding the encoded data is equal
		// to the original object
		if !apiequality.Semantic.DeepEqual(original, obj3) {
			t.Fatalf("%v: diff: %v\nCodec: %#v\nSource:\n\n%#v\n\nEncoded:\n\n%s\n\nFinal:\n\n%#v", name, diff.ObjectReflectDiff(original, obj3), codec, printer.Sprintf("%#v", original), dataAsString(data), printer.Sprintf("%#v", obj3))
		}
	}

}

// dataAsString returns the given byte array as a string; handles detecting
// protocol buffers.
func dataAsString(data []byte) string {
	dataString := string(data)
	if !strings.HasPrefix(dataString, "{") {
		dataString = "\n" + hex.Dump(data)
		proto.NewBuffer(make([]byte, 0, 1024)).DebugPrint("decoded object", data)
	}
	return dataString
}
