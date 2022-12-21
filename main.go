package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"k8s.io/klog/v2"

	"google.golang.org/protobuf/types/descriptorpb"
)

// func main() {
// 	ctx := context.Background()
// 	if err := run(ctx); err != nil {
// 		fmt.Fprintf(os.Stderr, "%v\n", err)
// 		os.Exit(1)
// 	}
// }

// func run(ctx context.Context) error {
// 	var g Generator

// 	p := "../api/apps/v1/"
// 	return v.VisitDir(p)
// }

func main() {
	var g Generator
	g.out = &ProtoWriter{w: os.Stdout}
	var analyzer = &analysis.Analyzer{
		Name: "genproto",
		Doc:  "generate proto output",
		Run:  g.Run,
	}
	singlechecker.Main(analyzer)
}

type Generator struct {
	*analysis.Pass
	out *ProtoWriter
}

func (g *Generator) Run(pass *analysis.Pass) (interface{}, error) {
	g.Pass = pass
	return nil, g.visitPass(pass)
}

func (g *Generator) visitPass(pass *analysis.Pass) error {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				switch decl.Tok {
				case token.TYPE:
					// klog.Infof("type %+v", decl)
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							if err := g.visitTypeSpec(spec); err != nil {
								return err
							}
						default:
							return fmt.Errorf("unhandled spec type %T", spec)
						}
					}
				case token.IMPORT:
					//klog.Infof("ast.Import")
				case token.VAR:
				//	klog.Infof("ast.Var")
				case token.CONST:
					//klog.Infof("ast.Const")
				default:
					return fmt.Errorf("unhandled GenDecl.Type=%v", decl.Tok)
				}
			case *ast.FuncDecl:
			default:
				return fmt.Errorf("unhandled type %T", decl)
			}
		}
	}
	return nil
}

// func (g *Generator) visitPass(pass *analysis.Pass) error {
// 	for k, def := range pass.TypesInfo.Defs {
// 		if err := g.visitTypeDef(k, def); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (g *Generator) visitTypeDef(k *ast.Ident, def types.Object) error {
// 	name := k.String()
// 	if def == nil {
// 		// ignore
// 		return nil
// 	}
// 	switch def := def.(type) {
// 	case *types.Var:
// 		// ignore
// 		return nil
// 	case *types.Func:
// 		// ignore
// 		return nil
// 	case *types.PkgName:
// 		// alias?  ignore
// 		return nil
// 	case *types.TypeName:
// 		// alias? ignore
// 		return nil
// 	case *types.Const:
// 		// ignore for now
// 		return nil
// 	default:
// 		return fmt.Errorf("unhandled typedef %q = %T", name, def)
// 	}
// }

// func (g *Generator) VisitDir(p string) error {
// 	var fset token.FileSet

// 	parsed, err := parser.ParseDir(&fset, p, nil, parser.AllErrors)
// 	if err != nil {
// 		return fmt.Errorf("error parsing %q: %w", p, err)
// 	}

// 	for _, pkg := range parsed {
// 		for _, file := range pkg.Files {
// 			for _, decl := range file.Decls {
// 				switch decl := decl.(type) {
// 				case *ast.GenDecl:
// 					switch decl.Tok {
// 					case token.TYPE:
// 						klog.Infof("type %+v", decl)
// 						for _, spec := range decl.Specs {
// 							switch spec := spec.(type) {
// 							case *ast.TypeSpec:
// 								if err := g.visitTypeSpec(spec); err != nil {
// 									return err
// 								}
// 							default:
// 								return fmt.Errorf("unhandled spec type %T", spec)
// 							}
// 						}
// 					case token.IMPORT:
// 						//klog.Infof("ast.Import")
// 					case token.VAR:
// 					//	klog.Infof("ast.Var")
// 					case token.CONST:
// 						//klog.Infof("ast.Const")
// 					default:
// 						return fmt.Errorf("unhandled GenDecl.Type=%v", decl.Tok)
// 					}
// 				case *ast.FuncDecl:
// 				default:
// 					return fmt.Errorf("unhandled type %T", decl)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func (g *Generator) visitTypeSpec(spec *ast.TypeSpec) error {
	// klog.Infof("  spec  %+v", spec)
	name := spec.Name.Name
	switch specType := spec.Type.(type) {
	case *ast.StructType:
		return g.visitStructType(name, specType)
	case *ast.MapType:
		// e.g. type ResourceList map[ResourceName]resource.Quantity
		// ignore
		return nil

	case *ast.SelectorExpr:
		// e.g. type List metav1.List
		// ignore
		return nil

	case *ast.Ident:
		return g.visitIdent(name, specType)
	default:
		return fmt.Errorf("unhandled TypeSpec::Type value %T (name=%q)", spec.Type, name)
	}
}

func (g *Generator) visitStructType(name string, spec *ast.StructType) error {

	// // ReplicaSetStatus represents the current status of a ReplicaSet.
	// message ReplicaSetStatus {
	// 	// Replicas is the most recently observed number of replicas.
	// 	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	// 	optional int32 replicas = 1;

	// 	// The number of pods that have labels matching the labels of the pod template of the replicaset.
	// 	// +optional
	// 	optional int32 fullyLabeledReplicas = 2;

	// 	// readyReplicas is the number of pods targeted by this ReplicaSet with a Ready Condition.
	// 	// +optional
	// 	optional int32 readyReplicas = 4;

	// 	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	// 	// +optional
	// 	optional int32 availableReplicas = 5;

	// 	// ObservedGeneration reflects the generation of the most recently observed ReplicaSet.
	// 	// +optional
	// 	optional int64 observedGeneration = 3;

	// 	// Represents the latest available observations of a replica set's current state.
	// 	// +optional
	// 	// +patchMergeKey=type
	// 	// +patchStrategy=merge
	// 	repeated ReplicaSetCondition conditions = 6;
	//   }

	msg := &descriptorpb.DescriptorProto{}
	msg.Name = PtrTo(name)

	// klog.Infof("  struct  %s", name)
	for _, field := range spec.Fields.List {
		name := ""
		if len(field.Names) == 0 {
			// TODO
		}
		if len(field.Names) > 1 {
			return fmt.Errorf("unexpected field with multiple names %v", field.Names)
		}
		for _, n := range field.Names {
			name = name + n.Name
		}

		f := &descriptorpb.FieldDescriptorProto{
			Name: &name,
		}
		if err := g.populateProtoFieldDescriptor(f, field.Type); err != nil {
			return err
		}

		if field.Tag != nil {
			if err := g.populateProtoFieldFromTag(f, field.Tag); err != nil {
				return err
			}
		}
		// klog.Infof("%s %s", field.Names, field.Type)
		msg.Field = append(msg.Field, f)
	}

	// klog.Infof("  msg  %s", prototext.Format(msg))

	g.out.WriteMessage(msg)
	return g.out.Err()
}

func (g *Generator) visitIdent(name string, spec *ast.Ident) error {
	klog.Infof("  ident %s %+v", name, spec)
	return nil
}

func (g *Generator) populateProtoFieldDescriptor(fd *descriptorpb.FieldDescriptorProto, fieldType ast.Expr) error {
	typeInfo, ok := g.TypesInfo.Types[fieldType]
	if !ok {
		return fmt.Errorf("no type info for %v", fieldType)
	}
	fd.Label = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()
	return g.populateProtoFieldDescriptorType(fd, typeInfo.Type)
}

func (g *Generator) populateProtoFieldDescriptorType(fd *descriptorpb.FieldDescriptorProto, typeInfo types.Type) error {
	typeName := ""
	switch typeInfo := typeInfo.(type) {
	case *types.Named:
		fd.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
		typeName = typeInfo.String()
	case *types.Basic:
		switch typeInfo.Kind() {
		case types.Bool:
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum()
		case types.Int32:
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum()
		case types.Int64:
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum()
		case types.String:
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
		default:
			return fmt.Errorf("unhandled basic kind %v", typeInfo.String())
		}
	case *types.Pointer:
		fd.Proto3Optional = PtrTo(true)
		if err := g.populateProtoFieldDescriptorType(fd, typeInfo.Elem()); err != nil {
			return err
		}
		return nil

	case *types.Slice:
		switch typeInfo.String() {
		case "[]byte":
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum()
			return nil
		default:
			fd.Label = descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum()
			if err := g.populateProtoFieldDescriptorType(fd, typeInfo.Elem()); err != nil {
				return err
			}
			return nil
		}

	case *types.Map:
		s := typeInfo.String()
		switch s {
		case "map[string]string":
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			typeName = "map<string, string>"

		case "map[string][]byte":
			fd.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			typeName = "map<string, bytes>"

		default:
			return fmt.Errorf("unsupported map type %q", s)
		}

	default:
		return fmt.Errorf("unhandled typeInfo.Type %T", typeInfo)
	}
	if typeName != "" {
		fd.TypeName = &typeName
	}

	return nil
}

func (g *Generator) populateProtoFieldFromTag(fd *descriptorpb.FieldDescriptorProto, tagLiteral *ast.BasicLit) error {
	tagValue := tagLiteral.Value
	if !strings.HasPrefix(tagValue, "`") || !strings.HasSuffix(tagValue, "`") {
		return fmt.Errorf("expected tags to be surrounded by `")
	}
	tagValue = strings.Trim(tagValue, "`")
	tags := strings.Fields(tagValue)
	for _, tag := range tags {
		if strings.HasPrefix(tag, "json:\"") {
			if !strings.HasSuffix(tag, "\"") {
				return fmt.Errorf("unimplemented tag %+v", tag)
			}
			tag = strings.TrimPrefix(tag, "json:\"")
			tag = strings.TrimSuffix(tag, "\"")
			omitEmpty := false
			if strings.HasSuffix(tag, ",omitempty") {
				omitEmpty = true
				tag = strings.TrimSuffix(tag, ",omitempty")
			}
			inline := false
			if strings.HasSuffix(tag, ",inline") {
				inline = true
				tag = strings.TrimSuffix(tag, ",inline")
			}
			tokens := strings.Split(tag, ",")
			if len(tokens) != 1 {
				return fmt.Errorf("unhandled json tag %q", tag)
			}
			klog.V(4).Infof("ignoring json tag %v, omitempty=%v, inline=%v", tag, omitEmpty, inline)
		} else if strings.HasPrefix(tag, "protobuf:\"") {
			if !strings.HasSuffix(tag, "\"") {
				return fmt.Errorf("unimplemented tag %+v", tag)
			}
			tag = strings.TrimPrefix(tag, "protobuf:\"")
			tag = strings.TrimSuffix(tag, "\"")
			tokens := strings.Split(tag, ",")
			for i, token := range tokens {
				if i == 0 {
					switch token {
					case "bytes":
						switch fd.GetType() {
						case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
							// ok
						case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
							// ok
						case descriptorpb.FieldDescriptorProto_TYPE_STRING:
							// ok
						case descriptorpb.FieldDescriptorProto_TYPE_INT32:
							klog.Warningf("changing type of int32 field to bytes: %v", formatProto(fd))
							fd.Type = descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum()

						case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
							klog.Warningf("changing type of bool field to bytes: %v", formatProto(fd))
							fd.Type = descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum()

						default:
							return fmt.Errorf("TODO: How do we specify bytes encoding for %v?", formatProto(fd))
						}
					case "varint":
						switch fd.GetType() {
						case descriptorpb.FieldDescriptorProto_TYPE_INT32:
							// ok
						case descriptorpb.FieldDescriptorProto_TYPE_INT64:
							// ok
						case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
							// ok
						default:
							return fmt.Errorf("TODO: How do we specify varint encoding for %v?", formatProto(fd))
						}
					default:
						return fmt.Errorf("unexpected protobuf tag %q", tag)
					}
				} else if i == 1 {
					number, err := strconv.Atoi(token)
					if err != nil {
						return fmt.Errorf("unexpected protobuf tag %q", tag)
					}
					fd.Number = PtrTo(int32(number))
				} else if token == "opt" {
					switch fd.GetLabel() {
					case descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL:
						// ok
					case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
						// ok
					default:
						return fmt.Errorf("protobuf tag was opt, but not marked as optional: %v", formatProto(fd))
					}
					// if ValueOf(fd.Proto3Optional) == false {
					// 	return fmt.Errorf("protobuf tag was optional, but not marked as optional: %v", formatProto(fd))
					// }
				} else if token == "rep" {
					switch fd.GetLabel() {
					case descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
						// ok
					default:
						klog.Warningf("protobuf tag was rep, but not marked as repeated: %v", formatProto(fd))
					}
				} else if strings.HasPrefix(token, "name=") {
					name := strings.TrimPrefix(token, "name=")
					fd.Name = &name
				} else {
					klog.Warningf("unknown protobuf tag %v", tokens)
				}
			}
		} else if strings.HasPrefix(tag, "patchStrategy:\"") {
			if !strings.HasSuffix(tag, "\"") {
				return fmt.Errorf("unimplemented tag %+v", tag)
			}
			tag = strings.TrimPrefix(tag, "patchStrategy:\"")
			tag = strings.TrimSuffix(tag, "\"")
			tokens := strings.Split(tag, ",")
			klog.Infof("ignoring patchStrategy tag %v", tokens)
		} else if strings.HasPrefix(tag, "patchMergeKey:\"") {
			if !strings.HasSuffix(tag, "\"") {
				return fmt.Errorf("unimplemented tag %+v", tag)
			}
			tag = strings.TrimPrefix(tag, "patchMergeKey:\"")
			tag = strings.TrimSuffix(tag, "\"")
			tokens := strings.Split(tag, ",")
			klog.Infof("ignoring patchMergeKey tag %v", tokens)
		} else {
			return fmt.Errorf("unimplemented tag %+v", tag)
		}
	}
	return nil
}

func PtrTo[T any](v T) *T {
	return &v
}

func ValueOf[T any](v *T) T {
	if v != nil {
		return *v
	}
	var def T
	return def
}
