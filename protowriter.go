package main

import (
	"bytes"
	"fmt"
	"io"

	"google.golang.org/protobuf/types/descriptorpb"
)

func formatProto(o interface{}) string {
	var b bytes.Buffer
	w := ProtoWriter{w: &b}
	switch o := o.(type) {
	case *descriptorpb.DescriptorProto:
		w.WriteMessage(o)
	case *descriptorpb.FieldDescriptorProto:
		w.writeField(&descriptorpb.DescriptorProto{}, o)
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

func (w *ProtoWriter) WriteHeader(packageName string, goPackageName string) {
	w.format("syntax = %q;\n", "proto3")
	w.format("\n")
	w.format("package %s;\n", packageName)
	w.format("\n")
	w.format("option go_package = %q;\n", goPackageName)
	w.format("\n")
}

func (w *ProtoWriter) WriteImport(p string) {
	w.format("import %q;\n", p)
}

func (w *ProtoWriter) WriteMessage(m *descriptorpb.DescriptorProto) {
	w.format("\n")
	w.format("message %s {\n", m.GetName())
	for _, field := range m.Field {
		w.writeField(m, field)
	}
	w.format("}\n")
}

func (w *ProtoWriter) writeField(msg *descriptorpb.DescriptorProto, fd *descriptorpb.FieldDescriptorProto) {
	// Check for map
	if fd.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		typeName := fd.GetTypeName()
		for _, nestedType := range msg.NestedType {
			if nestedType.GetName() == typeName {
				if nestedType.GetOptions().GetMapEntry() {
					w.writeMapField(msg, fd, nestedType)
					return
				}
			}
		}
	}

	w.format("  ")
	switch fd.GetLabel() {
	case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
		w.format("optional ")
	case descriptorpb.FieldDescriptorProto_LABEL_REQUIRED:
		w.format("required ")
	case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
		w.format("repeated ")
	default:
		w.error(fmt.Errorf("unexpected label %v", fd.GetLabel()))
	}
	w.writeType(fd)
	w.format(" %s = %d", fd.GetName(), fd.GetNumber())
	if fd.JsonName != nil {
		w.format(" [json_name = %q]", fd.GetJsonName())
	}
	w.format(";\n")
}

func (w *ProtoWriter) writeType(fd *descriptorpb.FieldDescriptorProto) {
	switch fd.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		w.format("string")
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		w.format("bool")
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		w.format("int32")
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		w.format("int64")
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		w.format("bytes")
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		w.format(fd.GetTypeName())
	default:
		w.error(fmt.Errorf("unexpected type %v", fd.GetType()))
	}
}

func (w *ProtoWriter) writeMapField(parent *descriptorpb.DescriptorProto, fd *descriptorpb.FieldDescriptorProto, mapType *descriptorpb.DescriptorProto) {
	w.format("  ")

	// optional is implicit
	// w.format("optional ")

	w.format("map<")
	var keyField *descriptorpb.FieldDescriptorProto
	var valueField *descriptorpb.FieldDescriptorProto
	for _, field := range mapType.Field {
		switch field.GetName() {
		case "key":
			keyField = field
		case "value":
			valueField = field
		default:
			w.error(fmt.Errorf("unexpected field %q in %q map", field.GetName(), fd.GetName()))
		}
	}
	if keyField == nil || valueField == nil {
		w.error(fmt.Errorf("missing key/value field in %q map", fd.GetName()))
	}

	w.writeType(keyField)
	w.format(", ")
	w.writeType(valueField)
	w.format("> ")

	w.format("%s = %d", fd.GetName(), fd.GetNumber())
	if fd.JsonName != nil {
		w.format(" [json_name = %q]", fd.GetJsonName())
	}
	w.format(";\n")
}
