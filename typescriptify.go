package tots

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tsDocTag = "ts_doc"
	tsType   = "ts_type"
)

// TypeOptions overrides options set by `ts_*` tags.
type TypeOptions struct {
	TSType string
	TSDoc  string
}

// StructType stores settings for transforming one Golang struct.
type StructType struct {
	Type         reflect.Type
	FieldOptions map[reflect.Type]TypeOptions
}

func NewStruct(i interface{}) *StructType {
	return &StructType{
		Type: reflect.TypeOf(i),
	}
}

func (st *StructType) WithFieldOpts(i interface{}, opts TypeOptions) *StructType {
	if st.FieldOptions == nil {
		st.FieldOptions = map[reflect.Type]TypeOptions{}
	}
	var typ reflect.Type
	if ty, is := i.(reflect.Type); is {
		typ = ty
	} else {
		typ = reflect.TypeOf(i)
	}
	st.FieldOptions[typ] = opts
	return st
}

type TypeScriptify struct {
	Prefix     string
	Suffix     string
	Indent     string
	DontExport bool

	structTypes []StructType
	kinds       map[reflect.Kind]string

	fieldTypeOptions map[reflect.Type]TypeOptions

	debug bool
	// throwaway, used when converting
	alreadyConverted map[reflect.Type]bool
}

func New() *TypeScriptify {
	result := new(TypeScriptify)
	result.Indent = "\t"

	kinds := make(map[reflect.Kind]string)

	kinds[reflect.Bool] = "boolean"

	kinds[reflect.Int] = "number"
	kinds[reflect.Int8] = "number"
	kinds[reflect.Int16] = "number"
	kinds[reflect.Int32] = "number"
	kinds[reflect.Int64] = "number"
	kinds[reflect.Uint] = "number"
	kinds[reflect.Uint8] = "number"
	kinds[reflect.Uint16] = "number"
	kinds[reflect.Uint32] = "number"
	kinds[reflect.Uint64] = "number"
	kinds[reflect.Float32] = "number"
	kinds[reflect.Float64] = "number"

	kinds[reflect.String] = "string"

	result.kinds = kinds

	result.Indent = "    "
	return result
}

func deepFields(typeOf reflect.Type) []reflect.StructField {
	fields := make([]reflect.StructField, 0)

	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	if typeOf.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < typeOf.NumField(); i++ {
		f := typeOf.Field(i)

		kind := f.Type.Kind()
		if f.Anonymous && kind == reflect.Struct {
			//fmt.Println(v.Interface())
			fields = append(fields, deepFields(f.Type)...)
		} else if f.Anonymous && kind == reflect.Ptr && f.Type.Elem().Kind() == reflect.Struct {
			//fmt.Println(v.Interface())
			fields = append(fields, deepFields(f.Type.Elem())...)
		} else {
			fields = append(fields, f)
		}
	}

	return fields
}

func (ts *TypeScriptify) Debug() *TypeScriptify {
	ts.debug = true
	return ts
}

func (ts *TypeScriptify) logf(depth int, s string, args ...interface{}) {
	if ts.debug {
		fmt.Printf(strings.Repeat("   ", depth)+s+"\n", args...)
	}
}

// ManageType can define custom options for fields of a specified type.
//
// This can be used instead of setting ts_type and ts_transform for all fields of a certain type.
func (t *TypeScriptify) ManageType(fld interface{}, opts TypeOptions) *TypeScriptify {
	var typ reflect.Type
	switch t := fld.(type) {
	case reflect.Type:
		typ = t
	default:
		typ = reflect.TypeOf(fld)
	}
	if t.fieldTypeOptions == nil {
		t.fieldTypeOptions = map[reflect.Type]TypeOptions{}
	}
	t.fieldTypeOptions[typ] = opts
	return t
}

func (t *TypeScriptify) WithIndent(i string) *TypeScriptify {
	t.Indent = i
	return t
}

func (t *TypeScriptify) WithPrefix(p string) *TypeScriptify {
	t.Prefix = p
	return t
}

func (t *TypeScriptify) WithSuffix(s string) *TypeScriptify {
	t.Suffix = s
	return t
}

func (t *TypeScriptify) Add(objs ...interface{}) *TypeScriptify {
	for _, obj := range objs {
		switch ty := obj.(type) {
		case StructType:
			t.structTypes = append(t.structTypes, ty)
		case *StructType:
			t.structTypes = append(t.structTypes, *ty)
		case reflect.Type:
			t.AddType(ty)
		default:
			t.AddType(reflect.TypeOf(obj))
		}
	}
	return t
}

func (t *TypeScriptify) AddType(typeOf reflect.Type) *TypeScriptify {
	t.structTypes = append(t.structTypes, StructType{Type: typeOf})
	return t
}

func (t *typeScriptClassBuilder) AddMapField(fieldName string, field reflect.StructField) {
	keyType := field.Type.Key()
	valueType := field.Type.Elem()
	valueTypeName := valueType.Name()
	if name, ok := t.types[valueType.Kind()]; ok {
		valueTypeName = name
	}
	if valueType.Kind() == reflect.Array || valueType.Kind() == reflect.Slice {
		valueTypeName = valueType.Elem().Name() + "[]"
	}
	if valueType.Kind() == reflect.Ptr {
		valueTypeName = valueType.Elem().Name()
	}
	keyTypeStr := keyType.Name()
	// Key should always be string, no need for this:
	// _, isSimple := t.types[keyType.Kind()]
	// if !isSimple {
	// 	keyTypeStr = t.prefix + keyType.Name() + t.suffix
	// }

	if valueType.Kind() == reflect.Struct {
		t.fields = append(t.fields, fmt.Sprintf("%s%s: {[key: %s]: %s};", t.indent, fieldName, keyTypeStr, t.prefix+valueTypeName))
	} else {
		t.fields = append(t.fields, fmt.Sprintf("%s%s: {[key: %s]: %s};", t.indent, fieldName, keyTypeStr, valueTypeName))
	}
}

func (t *TypeScriptify) Convert() (string, error) {
	t.alreadyConverted = make(map[reflect.Type]bool)
	depth := 0

	result := "/* Do not change, this code is generated from Golang structs */\n\n"
	for _, strctTyp := range t.structTypes {
		typeScriptCode, err := t.convertType(depth, strctTyp.Type)
		if err != nil {
			return "", err
		}
		result += "\n" + strings.Trim(typeScriptCode, " "+t.Indent+"\r\n")
	}
	return result, nil
}

type TSNamer interface {
	TSName() string
}

func (t *TypeScriptify) getFieldOptions(structType reflect.Type, field reflect.StructField) TypeOptions {
	// By default use options defined by tags:
	opts := TypeOptions{
		TSType: field.Tag.Get(tsType),
		TSDoc:  field.Tag.Get(tsDocTag),
	}

	overrides := []TypeOptions{}

	// But there is maybe an struct-specific override:
	for _, strct := range t.structTypes {
		if strct.FieldOptions == nil {
			continue
		}
		if strct.Type == structType {
			if fldOpts, found := strct.FieldOptions[field.Type]; found {
				overrides = append(overrides, fldOpts)
			}
		}
	}

	if fldOpts, found := t.fieldTypeOptions[field.Type]; found {
		overrides = append(overrides, fldOpts)
	}

	for _, o := range overrides {
		if o.TSType != "" {
			opts.TSType = o.TSType
		}
	}

	return opts
}

func (t *TypeScriptify) getJSONFieldName(field reflect.StructField) string {
	jsonFieldName := ""
	jsonTag := field.Tag.Get("json")
	if len(jsonTag) > 0 {
		jsonTagParts := strings.Split(jsonTag, ",")
		if len(jsonTagParts) > 0 {
			jsonFieldName = strings.Trim(jsonTagParts[0], t.Indent)
		}
		hasOmitEmpty := false
		for _, t := range jsonTagParts {
			if t == "" {
				break
			}
			if t == "omitempty" {
				hasOmitEmpty = true
				break
			}
		}
		if hasOmitEmpty {
			jsonFieldName = fmt.Sprintf("%s?", jsonFieldName)
		}
	} else if /*field.IsExported()*/ field.PkgPath == "" {
		jsonFieldName = field.Name
	}
	return jsonFieldName
}

func (t *TypeScriptify) convertType(depth int, typeOf reflect.Type) (string, error) {
	if _, found := t.alreadyConverted[typeOf]; found { // Already converted
		return "", nil
	}
	t.logf(depth, "Converting type %s", typeOf.String())

	t.alreadyConverted[typeOf] = true

	result := "{\n"
	builder := typeScriptClassBuilder{
		types:     t.kinds,
		indent:    t.Indent,
		prefix:    t.Prefix,
		suffix:    t.Suffix,
		typeIndex: -1,
	}

	fields := deepFields(typeOf)
	for _, field := range fields {
		isPtr := field.Type.Kind() == reflect.Ptr
		if isPtr {
			field.Type = field.Type.Elem()
		}
		jsonFieldName := t.getJSONFieldName(field)
		if len(jsonFieldName) == 0 || jsonFieldName == "-" {
			continue
		}

		var err error
		fldOpts := t.getFieldOptions(typeOf, field)
		if fldOpts.TSType != "" { // Struct:
			t.logf(depth, "- simple field %s.%s", typeOf.Name(), field.Name)
			err = builder.AddSimpleField(jsonFieldName, field, fldOpts)
		} else if field.Type.Kind() == reflect.Struct { // Struct:
			t.logf(depth, "- struct %s.%s (%s)", typeOf.Name(), field.Name, field.Type.String())
			typeScriptChunk, err := t.convertType(depth+1, field.Type)
			if err != nil {
				return "", err
			}
			if typeScriptChunk != "" {
				result = typeScriptChunk + "\n" + result
			}
			builder.AddStructField(jsonFieldName, field, fldOpts)
		} else if field.Type.Kind() == reflect.Map {
			t.logf(depth, "- map field %s.%s", typeOf.Name(), field.Name)
			// Also convert map key types if needed
			var keyTypeToConvert reflect.Type
			switch field.Type.Key().Kind() {
			case reflect.Struct:
				keyTypeToConvert = field.Type.Key()
			case reflect.Ptr:
				keyTypeToConvert = field.Type.Key().Elem()
			}
			if keyTypeToConvert != nil {
				typeScriptChunk, err := t.convertType(depth+1, keyTypeToConvert)
				if err != nil {
					return "", err
				}
				if typeScriptChunk != "" {
					result = typeScriptChunk + "\n" + result
				}
			}
			// Also convert map value types if needed
			var valueTypeToConvert reflect.Type
			switch field.Type.Elem().Kind() {
			case reflect.Struct:
				valueTypeToConvert = field.Type.Elem()
			case reflect.Ptr:
				valueTypeToConvert = field.Type.Elem().Elem()
			}
			if valueTypeToConvert != nil {
				typeScriptChunk, err := t.convertType(depth+1, valueTypeToConvert)
				if err != nil {
					return "", err
				}
				if typeScriptChunk != "" {
					result = typeScriptChunk + "\n" + result
				}
			}

			builder.AddMapField(jsonFieldName, field)
		} else if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array { // Slice:
			if field.Type.Elem().Kind() == reflect.Ptr { //extract ptr type
				field.Type = field.Type.Elem()
			}

			arrayDepth := 1
			for field.Type.Elem().Kind() == reflect.Slice { // Slice of slices:
				field.Type = field.Type.Elem()
				arrayDepth++
			}

			if field.Type.Elem().Kind() == reflect.Struct { // Slice of structs:
				t.logf(depth, "- struct slice %s.%s (%s)", typeOf.Name(), field.Name, field.Type.String())
				typeScriptChunk, err := t.convertType(depth+1, field.Type.Elem())
				if err != nil {
					return "", err
				}
				if typeScriptChunk != "" {
					result = typeScriptChunk + "\n" + result
				}
				builder.AddArrayOfStructsField(jsonFieldName, field, arrayDepth, fldOpts)
			} else { // Slice of simple fields:
				t.logf(depth, "- slice field %s.%s", typeOf.Name(), field.Name)
				err = builder.AddSimpleArrayField(jsonFieldName, field, arrayDepth, fldOpts)
			}
		} else { // Simple field:
			t.logf(depth, "- simple field %s.%s", typeOf.Name(), field.Name)
			err = builder.AddSimpleField(jsonFieldName, field, fldOpts)
		}
		if err != nil {
			return "", err
		}
	}

	entityName := t.Prefix + typeOf.Name() + t.Suffix
	result = fmt.Sprintf("interface %s%s ", entityName, builder.GetGenericType()) + result
	result += strings.Join(builder.fields, "\n") + "\n"
	result += "}"

	if !t.DontExport {
		result = "export " + result
	}
	return result, nil
}

type typeScriptClassBuilder struct {
	types          map[reflect.Kind]string
	indent         string
	fields         []string
	prefix, suffix string

	typeIndex int
}

func (t *typeScriptClassBuilder) GetGenericType() string {
	if t.typeIndex < 0 {
		return ""
	}

	ts := make([]string, 0, 5)
	for i := 0; i <= t.typeIndex; i++ {
		ts = append(ts, fmt.Sprintf("%c = any", byte('A'+i)))
	}
	return fmt.Sprintf("<%s>", strings.Join(ts, ", "))
}

func (t *typeScriptClassBuilder) AddSimpleArrayField(fieldName string, field reflect.StructField, arrayDepth int, opts TypeOptions) error {
	fieldType, kind := field.Type.Elem().Name(), field.Type.Elem().Kind()
	typeScriptType := t.types[kind]

	if len(fieldName) > 0 {
		if len(opts.TSType) > 0 {
			t.addField(fieldName, opts.TSType, opts)
			return nil
		} else if len(typeScriptType) > 0 {
			t.addField(fieldName, fmt.Sprint(typeScriptType, strings.Repeat("[]", arrayDepth)), opts)
			return nil
		}
	}

	return fmt.Errorf("cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType)
}

func (t *typeScriptClassBuilder) AddSimpleField(fieldName string, field reflect.StructField, opts TypeOptions) error {
	fieldType, kind := field.Type.Name(), field.Type.Kind()

	typeScriptType := t.types[kind]
	if len(opts.TSType) > 0 {
		typeScriptType = opts.TSType
	}

	if kind == reflect.Interface {
		t.typeIndex += 1
		typeScriptType = fmt.Sprintf("%c", byte('A'+t.typeIndex))
	}

	if len(typeScriptType) > 0 && len(fieldName) > 0 {
		t.addField(fieldName, typeScriptType, opts)
		return nil
	}

	return fmt.Errorf("cannot find type for %s (%s/%s)", kind.String(), fieldName, fieldType)
}

func (t *typeScriptClassBuilder) AddStructField(fieldName string, field reflect.StructField, opts TypeOptions) {
	fieldType := field.Type.Name()
	t.addField(fieldName, t.prefix+fieldType+t.suffix, opts)
}

func (t *typeScriptClassBuilder) AddArrayOfStructsField(fieldName string, field reflect.StructField, arrayDepth int, opts TypeOptions) {
	fieldType := field.Type.Elem().Name()
	t.addField(fieldName, fmt.Sprint(t.prefix+fieldType+t.suffix, strings.Repeat("[]", arrayDepth)), opts)
}

func (t *typeScriptClassBuilder) addField(fld, fldType string, opts TypeOptions) {
	doc := ""
	if opts.TSDoc != "" {
		doc = t.indent + "/**\n" +
			t.indent + " *\n" +
			t.indent + " * " + opts.TSDoc + "\n" +
			t.indent + " */\n"
	}
	t.fields = append(t.fields, fmt.Sprint(doc, t.indent, fld, ": ", fldType, ";"))
}
