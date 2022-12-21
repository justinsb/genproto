package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

func formatProto(o interface{}) string {
	var b bytes.Buffer
	w := ProtoWriter{w: &b}
	switch o := o.(type) {
	case *descriptorpb.DescriptorProto:
		w.WriteMessage(o)
	case *descriptorpb.FieldDescriptorProto:
		w.WriteField(o)
	default:
		return fmt.Sprintf("<unhandled type %T>", o)
	}
	if w.Err() != nil {
		return fmt.Sprintf("<error: %v>", w.Err())
	}
	return b.String()
}

type ProtoWriter struct {
	w   io.Writer
	err error
}

func (w *ProtoWriter) format(msg string, args ...any) {
	s := fmt.Sprintf(msg, args...)
	w.write(s)
}

func (w *ProtoWriter) write(s string) {
	if w.err != nil {
		return
	}

	if _, err := w.w.Write([]byte(s)); err != nil {
		w.err = err
	}
}

func (w *ProtoWriter) error(err error) {
	if w.err == nil {
		w.err = err
	}
}

func (w *ProtoWriter) Err() error {
	return w.err
}

func (w *ProtoWriter) WriteMessage(m *descriptorpb.DescriptorProto) {
	w.format("message %s {\n", m.GetName())
	for _, field := range m.Field {
		w.WriteField(field)
	}
	w.format("}\n")
}

func (w *ProtoWriter) WriteField(m *descriptorpb.FieldDescriptorProto) {
	var s strings.Builder
	s.WriteString("  ")
	switch m.GetLabel() {
	case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
		s.WriteString("optional ")
	case descriptorpb.FieldDescriptorProto_LABEL_REQUIRED:
		s.WriteString("required ")
	case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
		s.WriteString("repeated ")
	default:
		w.error(fmt.Errorf("unexpected label %v", m.GetLabel()))
	}
	switch m.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		s.WriteString("string ")
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		s.WriteString("bool ")
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		s.WriteString("double ")
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		s.WriteString("int32 ")
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		s.WriteString("int64 ")
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		s.WriteString("bytes ")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		s.WriteString(m.GetTypeName() + " ")
	default:
		w.error(fmt.Errorf("unexpected type %v", m.GetType()))
	}
	w.write(s.String())
	w.format("%s = %d;\n", m.GetName(), m.GetNumber())
}
