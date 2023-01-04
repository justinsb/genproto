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
	"time"

	kubeemetav1 "justinsb.com/kubee/apimachinery/pkg/apis/meta/v1"
	kubeeruntime "justinsb.com/kubee/apimachinery/pkg/runtime"
	extensionsv1 "justinsb.com/kubee/kubee/v1"
	"k8s.io/klog/v2"

	"github.com/davecgh/go-spew/spew"
	legacyproto "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	genericfuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	kubeproto "k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/apimachinery/pkg/util/diff"

	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	admissionregv1alpha1 "k8s.io/api/admissionregistration/v1alpha1"
	admissionregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apidiscoveryv2beta1 "k8s.io/api/apidiscovery/v2beta1"
	apiserverinternalv1alpha1 "k8s.io/api/apiserverinternal/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	authenticationv1 "k8s.io/api/authentication/v1"
	authenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	authorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	coordinationv1 "k8s.io/api/coordination/v1"
	coordinationv1beta1 "k8s.io/api/coordination/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	eventsv1 "k8s.io/api/events/v1"
	eventsv1beta1 "k8s.io/api/events/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	flowcontrolv1alpha1 "k8s.io/api/flowcontrol/v1alpha1"
	flowcontrolv1beta1 "k8s.io/api/flowcontrol/v1beta1"
	flowcontrolv1beta2 "k8s.io/api/flowcontrol/v1beta2"
	flowcontrolv1beta3 "k8s.io/api/flowcontrol/v1beta3"
	imagepolicyv1alpha1 "k8s.io/api/imagepolicy/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1alpha1 "k8s.io/api/networking/v1alpha1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	nodev1 "k8s.io/api/node/v1"
	nodev1alpha1 "k8s.io/api/node/v1alpha1"
	nodev1beta1 "k8s.io/api/node/v1beta1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	resourcev1alpha1 "k8s.io/api/resource/v1alpha1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	schedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"

	kubeeadmissionv1 "justinsb.com/kubee/api/admission/v1"
	kubeeadmissionv1beta1 "justinsb.com/kubee/api/admission/v1beta1"
	kubeeadmissionregv1 "justinsb.com/kubee/api/admissionregistration/v1"
	kubeeadmissionregv1alpha1 "justinsb.com/kubee/api/admissionregistration/v1alpha1"
	kubeeadmissionregv1beta1 "justinsb.com/kubee/api/admissionregistration/v1beta1"
	kubeeapidiscoveryv2beta1 "justinsb.com/kubee/api/apidiscovery/v2beta1"
	kubeeapiserverinternalv1alpha1 "justinsb.com/kubee/api/apiserverinternal/v1alpha1"
	kubeeappsv1 "justinsb.com/kubee/api/apps/v1"
	kubeeappsv1beta1 "justinsb.com/kubee/api/apps/v1beta1"
	kubeeappsv1beta2 "justinsb.com/kubee/api/apps/v1beta2"
	kubeeauthenticationv1 "justinsb.com/kubee/api/authentication/v1"
	kubeeauthenticationv1beta1 "justinsb.com/kubee/api/authentication/v1beta1"
	kubeeauthorizationv1 "justinsb.com/kubee/api/authorization/v1"
	kubeeauthorizationv1beta1 "justinsb.com/kubee/api/authorization/v1beta1"
	kubeeautoscalingv1 "justinsb.com/kubee/api/autoscaling/v1"
	kubeeautoscalingv2 "justinsb.com/kubee/api/autoscaling/v2"
	kubeeautoscalingv2beta1 "justinsb.com/kubee/api/autoscaling/v2beta1"
	kubeeautoscalingv2beta2 "justinsb.com/kubee/api/autoscaling/v2beta2"
	kubeebatchv1 "justinsb.com/kubee/api/batch/v1"
	kubeebatchv1beta1 "justinsb.com/kubee/api/batch/v1beta1"
	kubeecertificatesv1 "justinsb.com/kubee/api/certificates/v1"
	kubeecertificatesv1beta1 "justinsb.com/kubee/api/certificates/v1beta1"
	kubeecoordinationv1 "justinsb.com/kubee/api/coordination/v1"
	kubeecoordinationv1beta1 "justinsb.com/kubee/api/coordination/v1beta1"
	kubeecorev1 "justinsb.com/kubee/api/core/v1"
	kubeediscoveryv1 "justinsb.com/kubee/api/discovery/v1"
	kubeediscoveryv1beta1 "justinsb.com/kubee/api/discovery/v1beta1"
	kubeeeventsv1 "justinsb.com/kubee/api/events/v1"
	kubeeeventsv1beta1 "justinsb.com/kubee/api/events/v1beta1"
	kubeeextensionsv1beta1 "justinsb.com/kubee/api/extensions/v1beta1"
	kubeeflowcontrolv1alpha1 "justinsb.com/kubee/api/flowcontrol/v1alpha1"
	kubeeflowcontrolv1beta1 "justinsb.com/kubee/api/flowcontrol/v1beta1"
	kubeeflowcontrolv1beta2 "justinsb.com/kubee/api/flowcontrol/v1beta2"
	kubeeflowcontrolv1beta3 "justinsb.com/kubee/api/flowcontrol/v1beta3"
	kubeeimagepolicyv1alpha1 "justinsb.com/kubee/api/imagepolicy/v1alpha1"
	kubeenetworkingv1 "justinsb.com/kubee/api/networking/v1"
	kubeenetworkingv1alpha1 "justinsb.com/kubee/api/networking/v1alpha1"
	kubeenetworkingv1beta1 "justinsb.com/kubee/api/networking/v1beta1"
	kubeenodev1 "justinsb.com/kubee/api/node/v1"
	kubeenodev1alpha1 "justinsb.com/kubee/api/node/v1alpha1"
	kubeenodev1beta1 "justinsb.com/kubee/api/node/v1beta1"
	kubeepolicyv1 "justinsb.com/kubee/api/policy/v1"
	kubeepolicyv1beta1 "justinsb.com/kubee/api/policy/v1beta1"
	kubeerbacv1 "justinsb.com/kubee/api/rbac/v1"
	kubeerbacv1alpha1 "justinsb.com/kubee/api/rbac/v1alpha1"
	kubeerbacv1beta1 "justinsb.com/kubee/api/rbac/v1beta1"
	kubeeresourcev1alpha1 "justinsb.com/kubee/api/resource/v1alpha1"
	kubeeschedulingv1 "justinsb.com/kubee/api/scheduling/v1"
	kubeeschedulingv1alpha1 "justinsb.com/kubee/api/scheduling/v1alpha1"
	kubeeschedulingv1beta1 "justinsb.com/kubee/api/scheduling/v1beta1"
	kubeestoragev1 "justinsb.com/kubee/api/storage/v1"
	kubeestoragev1alpha1 "justinsb.com/kubee/api/storage/v1alpha1"
	kubeestoragev1beta1 "justinsb.com/kubee/api/storage/v1beta1"
)

var groups = []runtime.SchemeBuilder{
	admissionv1beta1.SchemeBuilder,
	admissionv1.SchemeBuilder,
	admissionregv1alpha1.SchemeBuilder,
	admissionregv1beta1.SchemeBuilder,
	admissionregv1.SchemeBuilder,
	apiserverinternalv1alpha1.SchemeBuilder,
	apidiscoveryv2beta1.SchemeBuilder,
	appsv1beta1.SchemeBuilder,
	appsv1beta2.SchemeBuilder,
	appsv1.SchemeBuilder,
	authenticationv1beta1.SchemeBuilder,
	authenticationv1.SchemeBuilder,
	authorizationv1beta1.SchemeBuilder,
	authorizationv1.SchemeBuilder,
	autoscalingv1.SchemeBuilder,
	autoscalingv2.SchemeBuilder,
	autoscalingv2beta1.SchemeBuilder,
	autoscalingv2beta2.SchemeBuilder,
	batchv1beta1.SchemeBuilder,
	batchv1.SchemeBuilder,
	certificatesv1.SchemeBuilder,
	certificatesv1beta1.SchemeBuilder,
	coordinationv1.SchemeBuilder,
	coordinationv1beta1.SchemeBuilder,
	corev1.SchemeBuilder,
	discoveryv1.SchemeBuilder,
	discoveryv1beta1.SchemeBuilder,
	eventsv1.SchemeBuilder,
	eventsv1beta1.SchemeBuilder,
	extensionsv1beta1.SchemeBuilder,
	flowcontrolv1alpha1.SchemeBuilder,
	flowcontrolv1beta1.SchemeBuilder,
	flowcontrolv1beta2.SchemeBuilder,
	flowcontrolv1beta3.SchemeBuilder,
	imagepolicyv1alpha1.SchemeBuilder,
	networkingv1.SchemeBuilder,
	networkingv1beta1.SchemeBuilder,
	networkingv1alpha1.SchemeBuilder,
	nodev1.SchemeBuilder,
	nodev1alpha1.SchemeBuilder,
	nodev1beta1.SchemeBuilder,
	policyv1.SchemeBuilder,
	policyv1beta1.SchemeBuilder,
	rbacv1alpha1.SchemeBuilder,
	rbacv1beta1.SchemeBuilder,
	rbacv1.SchemeBuilder,
	resourcev1alpha1.SchemeBuilder,
	schedulingv1alpha1.SchemeBuilder,
	schedulingv1beta1.SchemeBuilder,
	schedulingv1.SchemeBuilder,
	storagev1alpha1.SchemeBuilder,
	storagev1beta1.SchemeBuilder,
	storagev1.SchemeBuilder,
}

var kubeeGroups = []protoreflect.FileDescriptor{
	kubeeadmissionv1beta1.File_api_admission_v1beta1_generated_proto,
	kubeeadmissionv1.File_api_admission_v1_generated_proto,
	kubeeadmissionregv1alpha1.File_api_admissionregistration_v1alpha1_generated_proto,
	kubeeadmissionregv1beta1.File_api_admissionregistration_v1beta1_generated_proto,
	kubeeadmissionregv1.File_api_admissionregistration_v1_generated_proto,
	kubeeapiserverinternalv1alpha1.File_api_apiserverinternal_v1alpha1_generated_proto,
	kubeeapidiscoveryv2beta1.File_api_apidiscovery_v2beta1_generated_proto,
	kubeeappsv1beta1.File_api_apps_v1beta1_generated_proto,
	kubeeappsv1beta2.File_api_apps_v1beta2_generated_proto,
	kubeeappsv1.File_api_apps_v1_generated_proto,
	kubeeauthenticationv1beta1.File_api_authentication_v1beta1_generated_proto,
	kubeeauthenticationv1.File_api_authentication_v1_generated_proto,
	kubeeauthorizationv1beta1.File_api_authorization_v1beta1_generated_proto,
	kubeeauthorizationv1.File_api_authorization_v1_generated_proto,
	kubeeautoscalingv1.File_api_autoscaling_v1_generated_proto,
	kubeeautoscalingv2.File_api_autoscaling_v2_generated_proto,
	kubeeautoscalingv2beta1.File_api_autoscaling_v2beta1_generated_proto,
	kubeeautoscalingv2beta2.File_api_autoscaling_v2beta2_generated_proto,
	kubeebatchv1beta1.File_api_batch_v1beta1_generated_proto,
	kubeebatchv1.File_api_batch_v1_generated_proto,
	kubeecertificatesv1.File_api_certificates_v1_generated_proto,
	kubeecertificatesv1beta1.File_api_certificates_v1beta1_generated_proto,
	kubeecoordinationv1.File_api_coordination_v1_generated_proto,
	kubeecoordinationv1beta1.File_api_coordination_v1beta1_generated_proto,
	kubeecorev1.File_api_core_v1_generated_proto,
	kubeediscoveryv1.File_api_discovery_v1_generated_proto,
	kubeediscoveryv1beta1.File_api_discovery_v1beta1_generated_proto,
	kubeeeventsv1.File_api_events_v1_generated_proto,
	kubeeeventsv1beta1.File_api_events_v1beta1_generated_proto,
	kubeeextensionsv1beta1.File_api_extensions_v1beta1_generated_proto,
	kubeeflowcontrolv1alpha1.File_api_flowcontrol_v1alpha1_generated_proto,
	kubeeflowcontrolv1beta1.File_api_flowcontrol_v1beta1_generated_proto,
	kubeeflowcontrolv1beta2.File_api_flowcontrol_v1beta2_generated_proto,
	kubeeflowcontrolv1beta3.File_api_flowcontrol_v1beta3_generated_proto,
	kubeeimagepolicyv1alpha1.File_api_imagepolicy_v1alpha1_generated_proto,
	kubeenetworkingv1.File_api_networking_v1_generated_proto,
	kubeenetworkingv1beta1.File_api_networking_v1beta1_generated_proto,
	kubeenetworkingv1alpha1.File_api_networking_v1alpha1_generated_proto,
	kubeenodev1.File_api_node_v1_generated_proto,
	kubeenodev1alpha1.File_api_node_v1alpha1_generated_proto,
	kubeenodev1beta1.File_api_node_v1beta1_generated_proto,
	kubeepolicyv1.File_api_policy_v1_generated_proto,
	kubeepolicyv1beta1.File_api_policy_v1beta1_generated_proto,
	kubeerbacv1alpha1.File_api_rbac_v1alpha1_generated_proto,
	kubeerbacv1beta1.File_api_rbac_v1beta1_generated_proto,
	kubeerbacv1.File_api_rbac_v1_generated_proto,
	kubeeresourcev1alpha1.File_api_resource_v1alpha1_generated_proto,
	kubeeschedulingv1alpha1.File_api_scheduling_v1alpha1_generated_proto,
	kubeeschedulingv1beta1.File_api_scheduling_v1beta1_generated_proto,
	kubeeschedulingv1.File_api_scheduling_v1_generated_proto,
	kubeestoragev1alpha1.File_api_storage_v1alpha1_generated_proto,
	kubeestoragev1beta1.File_api_storage_v1beta1_generated_proto,
	kubeestoragev1.File_api_storage_v1_generated_proto,
}

func TestRoundTrip(t *testing.T) {
	h := NewFuzzHarness(t)

	rand.Seed(time.Now().UnixNano())

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

	kubeeSchema SchemaInfo
	scheme      *runtime.Scheme
	codecs      serializer.CodecFactory
	encodings   map[string]runtime.Codec
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

	h := &FuzzHarness{
		T:         t,
		codecs:    codecs,
		scheme:    scheme,
		encodings: encodings,
	}

	h.kubeeSchema.RegisterFileDescriptor(kubeemetav1.File_apimachinery_pkg_apis_meta_v1_generated_proto)

	for _, kubeeGroup := range kubeeGroups {
		h.kubeeSchema.RegisterFileDescriptor(kubeeGroup)
	}

	return h
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

type SchemaInfo struct {
	kindsByGVK map[GroupVersionKind]*KindInfo
}

type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

type KindInfo struct {
	GroupVersionKind
	Resource string

	messageType protoreflect.MessageType
}

func (k *KindInfo) New() protoreflect.Message {
	return k.messageType.New()
}

func (s *SchemaInfo) KindByGVK(gvk GroupVersionKind) *KindInfo {
	return s.kindsByGVK[gvk]
}

func (s *SchemaInfo) RegisterFileDescriptor(fd protoreflect.FileDescriptor) {
	fileOptions := fd.Options().(*descriptorpb.FileOptions)
	groupVersionVal := fileOptions.ProtoReflect().Get(extensionsv1.E_GroupVersion.TypeDescriptor())
	groupVersion, ok := groupVersionVal.Message().Interface().(*extensionsv1.GroupVersion)
	if !ok {
		klog.Fatalf("unexpected type for group_version annotation, got %T", groupVersionVal.Message().Interface())
	}

	n := fd.Messages().Len()
	for i := 0; i < n; i++ {
		msg := fd.Messages().Get(i)
		s.registerMessage(groupVersion, msg)
	}
}

func (s *SchemaInfo) registerMessage(groupVersion *extensionsv1.GroupVersion, msg protoreflect.MessageDescriptor) {
	messageOptions := msg.Options().(*descriptorpb.MessageOptions)
	kindVal := messageOptions.ProtoReflect().Get(extensionsv1.E_Kind.TypeDescriptor())

	kind, ok := kindVal.Message().Interface().(*extensionsv1.Kind)
	if !ok {
		klog.Fatalf("unexpected type for kind annotation, got %T", kindVal.Message().Interface())
	}

	if kind == nil {
		return
	}

	messageType, err := protoregistry.GlobalTypes.FindMessageByName(msg.FullName())
	if err != nil {
		klog.Fatalf("failed to find message %q: %v", msg.FullName(), err)
	}

	info := &KindInfo{
		messageType: messageType,
	}
	info.Kind = kind.Kind
	if info.Kind == "" {
		info.Kind = string(msg.Name())
	}

	info.Resource = kind.Resource
	if info.Resource == "" {
		resource := info.Kind
		resource = strings.ToLower(resource)
		resource += "s" // TODO: worry about pluralization rules?
		info.Resource = resource
	}

	info.Group = groupVersion.Group
	info.Version = groupVersion.Version
	if info.Group == "" {
		// This is likely "core"
	}
	if info.Version == "" {
		// Default from path?
		klog.Fatalf("group_version not found; unable to determine version for %v", string(msg.Name()))
	}

	if s.kindsByGVK == nil {
		s.kindsByGVK = make(map[GroupVersionKind]*KindInfo)
	}
	s.kindsByGVK[info.GroupVersionKind] = info
}

func (t *FuzzHarness) testRoundTrip(codec runtime.Codec, object runtime.Object, unmarshal func([]byte, interface{}) error, marshal func(obj interface{}) ([]byte, error)) {
	printer := spew.ConfigState{DisableMethods: true}
	original := object

	data := t.encodeWithCodec(codec, object)

	object = object.DeepCopyObject()

	var kubee interface{}
	gvk := object.GetObjectKind().GroupVersionKind()

	kubeeGVK := GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
	kind := t.kubeeSchema.KindByGVK(kubeeGVK)
	if kind == nil {
		switch kubeeGVK.Kind {
		case "ListOptions", "PatchOptions", "UpdateOptions", "CreateOptions", "DeleteOptions", "GetOptions":
			// TODO: Do we need these in each group?
			t.Skip("skipping options round-trip test")
		case "Status":
			if kubeeGVK.Group == "resource.k8s.io" && kubeeGVK.Version == "v1alpha1" {
				// TODO: What is this?
				t.Skip("skipping resource.k8s.io.v1alpha1.Status round-trip test")
			}
		}
		t.Fatalf("kind %v not known", kubeeGVK)
	}
	kubee = kind.New().Interface()

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
