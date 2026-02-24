package libxdr

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/groundhogchain/lib-go/must"
	"github.com/xdrpp/goxdr/xdr"
)

type jsonIn struct {
	obj any
}

func mustString(i any) string {
	switch i.(type) {
	case map[string]any, []any:
		xdr.XdrPanic("JsonToXdr: Expected scalar got %T", i)
	}
	return fmt.Sprint(i)
}

func (j *jsonIn) get(name string) any {
	if len(name) == 0 {
		return j.obj
	} else if name[0] == '[' {
		a := j.obj.([]any)
		if len(a) == 0 {
			xdr.XdrPanic("JsonToXdr: insufficient elements in array")
		}
		j.obj = a[1:]
		return a[0]
	} else if obj, ok := j.obj.(map[string]any)[name]; ok {
		return obj
	}
	return nil
}

func (_ *jsonIn) Sprintf(f string, args ...any) string {
	return fmt.Sprintf(f, args...)
}
func (j *jsonIn) Marshal(name string, xval xdr.XdrType) {
	jval := j.get(name)
	if HideFieldName(name, xval) {
		jval = j.obj
	}
	if jval == nil {
		return
	}
	switch v := xdr.XdrBaseType(xval).(type) {
	case xdr.XdrString:
		v.SetString(mustString(jval))
	case xdr.XdrVecOpaque:
		bs, err := hex.DecodeString(mustString(jval))
		if err != nil {
			panic(err)
		}
		v.SetByteSlice(bs)
	case xdr.XdrArrayOpaque:
		dst := v.GetByteSlice()
		bs, err := hex.DecodeString(mustString(jval))
		if err != nil {
			panic(err)
		} else if len(bs) != len(dst) {
			xdr.XdrPanic("JsonToXdr: %s decodes to %d bytes, want %d bytes",
				name, len(bs), len(dst))
		}
		copy(dst, bs)
	case fmt.Scanner:
		if _, err := fmt.Sscan(mustString(jval), v); err != nil {
			xdr.XdrPanic("JsonToXdr: %s: %s", name, err.Error())
		}
	case xdr.XdrPtr:
		v.SetPresent(true)
		v.XdrMarshalValue(j, name)
	case xdr.XdrVec:
		v.XdrMarshalN(&jsonIn{jval}, "", uint32(len(jval.([]any))))
	case xdr.XdrUnion:
		j.marshalUnion(v, jval)
	case xdr.XdrAggregate:
		v.XdrRecurse(&jsonIn{jval}, "")
	}
}

func fieldFromXdrName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func (j *jsonIn) marshalUnion(u xdr.XdrUnion, jval any) {
	// Back-compat: accept legacy flat union JSON that includes the
	// discriminant field (e.g. {"scheme":"Ed25519","ed25519":"..."}).
	tagField := fieldFromXdrName(u.XdrUnionTagName())
	if jm, ok := jval.(map[string]any); ok {
		if _, hasLegacyTag := jm[tagField]; !hasLegacyTag {
			// New tagged format:
			// {"Ed25519":{"ed25519":"..."}}
			if len(jm) != 1 {
				xdr.XdrPanic("JsonToXdr: expected exactly one union arm for %s", u.XdrTypeName())
			}
			var tagText string
			var bodyVal any
			for k, v := range jm {
				tagText, bodyVal = k, v
			}
			if _, err := fmt.Sscan(tagText, u.XdrUnionTag()); err != nil {
				xdr.XdrPanic("JsonToXdr: invalid union arm %q for %s", tagText, u.XdrTypeName())
			}
			body := u.XdrUnionBody()
			if body == nil {
				return
			}
			bodyField := fieldFromXdrName(u.XdrUnionBodyName())
			switch m := bodyVal.(type) {
			case map[string]any:
				if _, ok := m[bodyField]; ok {
					(&jsonIn{obj: m}).Marshal(bodyField, body)
				} else if len(m) == 1 {
					// Tolerate wrappers that use a different field spelling.
					for _, v := range m {
						(&jsonIn{obj: map[string]any{bodyField: v}}).Marshal(bodyField, body)
					}
				} else {
					xdr.XdrPanic("JsonToXdr: malformed union body for %s", u.XdrTypeName())
				}
			default:
				// Tolerate shorthand {"Arm": <value>}.
				(&jsonIn{obj: map[string]any{bodyField: bodyVal}}).Marshal(bodyField, body)
			}
			return
		}
	}
	u.XdrRecurse(&jsonIn{jval}, "")
}

// Parse JSON into an XDR structure.  This function assumes that dst
// is in a pristine state--for example, it won't reset pointers to nil
// if the corresponding JSON is null.  Moreover, it is somewhat strict
// in expecting the JSON to conform to the XDR.
func JsonToXdr(dst xdr.XdrType, src []byte) (err error) {
	defer func() {
		if i := recover(); i != nil {
			err = i.(error)
		}
	}()
	var j jsonIn
	json.Unmarshal(src, &j.obj)
	j.Marshal("", dst)
	return nil
}

type jsonOut struct {
	out       *bytes.Buffer
	indent    string
	needComma bool
}

func (j *jsonOut) printField(name string, f string, args ...any) {
	if j.needComma {
		j.out.WriteString(",\n")
	} else {
		j.out.WriteString("\n")
		j.needComma = true
	}
	j.out.WriteString(j.indent)
	if len(name) > 0 && name[0] != '[' {
		fmt.Fprintf(j.out, "%q: ", name)
	}
	fmt.Fprintf(j.out, f, args...)
}

func (j *jsonOut) aggregate(val xdr.XdrAggregate) {
	oldIndent := j.indent
	defer func() {
		j.needComma = true
		j.indent = oldIndent
	}()
	j.indent = j.indent + "    "
	j.needComma = false
	switch v := val.(type) {
	case xdr.XdrVec:
		j.out.WriteString("[")
		v.XdrMarshalN(j, "", v.GetVecLen())
		j.out.WriteString("\n" + oldIndent + "]")
	case xdr.XdrArray:
		j.out.WriteString("[")
		v.XdrRecurse(j, "")
		j.out.WriteString("\n" + oldIndent + "]")
	case xdr.XdrAggregate:
		j.out.WriteString("{")
		v.XdrRecurse(j, "")
		j.out.WriteString("\n" + oldIndent + "}")
	}
}

func (_ *jsonOut) Sprintf(f string, args ...any) string {
	return fmt.Sprintf(f, args...)
}
func (j *jsonOut) Marshal(name string, val xdr.XdrType) {
	switch v := xdr.XdrBaseType(val).(type) {
	case *xdr.XdrBool:
		j.printField(name, "%v", *v)
	case xdr.XdrEnum:
		j.printField(name, "%q", v.String())
	case xdr.XdrNum32:
		j.printField(name, "%s", v.String())
		// Intentionally don't do the same for 64-bit, which gets
		// passed as strings to avoid any loss of precision.
	case xdr.XdrString:
		j.printField(name, "%s", v.String())
	case xdr.XdrBytes:
		j.printField(name, "\"")
		j.out.WriteString(hex.EncodeToString(v.GetByteSlice()))
		j.out.WriteByte('"')
	case fmt.Stringer:
		j.printField(name, "%q", v.String())
	case xdr.XdrPtr:
		if !v.GetPresent() {
			j.printField(name, "null")
		} else {
			v.XdrMarshalValue(j, name)
		}
	case xdr.XdrUnion:
		j.marshalUnion(name, v)
	case xdr.XdrAggregate:
		if HideFieldName(name, val) {
			v.XdrRecurse(j, "")
		} else {
			j.printField(name, "")
			j.aggregate(v)
		}
	case *xdrGo:
		j.marshalGo(name, v.Value)
	default:
		xdr.XdrPanic("XdrToJson can't handle type %T", val)
	}
}

func (j *jsonOut) marshalUnion(name string, u xdr.XdrUnion) {
	j.printField(name, "")
	j.out.WriteString("{")
	armName := u.XdrUnionTag().String()
	body := u.XdrUnionBody()
	nj := &jsonOut{
		out:       j.out,
		indent:    j.indent + "    ",
		needComma: false,
	}
	if body == nil {
		nj.printField(armName, "null")
	} else {
		bodyField := fieldFromXdrName(u.XdrUnionBodyName())
		nj.printField(armName, "")
		nj.out.WriteString("{")
		bj := &jsonOut{
			out:       nj.out,
			indent:    nj.indent + "    ",
			needComma: false,
		}
		bj.Marshal(bodyField, body)
		nj.out.WriteString("\n" + nj.indent + "}")
	}
	j.out.WriteString("\n" + j.indent + "}")
}

func (j *jsonOut) marshalGo(name string, val any) {
	switch u := (val).(type) {
	case bool:
		x := xdr.XdrBool(u)
		j.Marshal(name, &x)
	case int32:
		x := xdr.XdrInt32(u)
		j.Marshal(name, &x)
	case uint32:
		x := xdr.XdrUint32(u)
		j.Marshal(name, &x)
	case int64:
		x := xdr.XdrInt64(u)
		j.Marshal(name, &x)
	case uint64:
		x := xdr.XdrUint64(u)
		j.Marshal(name, &x)
	case float32:
		x := xdr.XdrFloat32(u)
		j.Marshal(name, &x)
	case float64:
		x := xdr.XdrFloat64(u)
		j.Marshal(name, &x)
	case string:
		x := xdr.XdrString{Str: &u, Bound: uint32(len(u))}
		j.Marshal(name, x)
	case []byte:
		x := xdr.XdrVecOpaque{Bytes: &u, Bound: uint32(len(u))}
		j.Marshal(name, &x)
	case time.Time:
		j.printField(name, "%v", u)
	default:
		vv := reflect.ValueOf(val)
		switch {
		case vv.Kind() == reflect.Array || vv.Kind() == reflect.Slice:
			switch {
			case vv.Type().Elem() == byteType: // []byte, [n]byte
				j.printField(name, "\"")
				aa := reflect.New(vv.Type()).Elem()
				aa.Set(vv)
				j.out.Write([]byte(hex.EncodeToString(aa.Slice(0, aa.Len()).Bytes())))
				j.out.WriteByte('"')
			default: // []T, [n]T
				j.printField(name, "")
				j.aggregateGo(vv)
			}
		case vv.Kind() == reflect.Struct:
			j.printField(name, "")
			j.aggregateGo(vv)
		case vv.Kind() == reflect.Interface || vv.Kind() == reflect.Pointer:
			if vv.IsNil() {
				j.printField(name, "<nil>")
			} else {
				j.marshalGo(name, vv.Elem().Interface())
			}
		default: // complex64, etc
			j.printField(name, "%v", val)
		}
	}
}

func (j *jsonOut) aggregateGo(val reflect.Value) {
	oldIndent := j.indent
	defer func() {
		j.needComma = true
		j.indent = oldIndent
	}()
	j.indent = j.indent + "    "
	j.needComma = false
	switch {
	case val.Type().Kind() == reflect.Slice:
		j.out.WriteString("[")
		for i := 0; i < val.Len(); i++ {
			j.marshalGo("", val.Index(i).Interface())
		}
		j.out.WriteString("\n" + oldIndent + "]")
	case val.Type().Kind() == reflect.Array:
		j.out.WriteString("[")
		for i := 0; i < val.Len(); i++ {
			j.marshalGo("", val.Index(i).Interface())
		}
		j.out.WriteString("\n" + oldIndent + "]")
	case val.Type().Kind() == reflect.Struct:
		j.out.WriteString("{")
		for i := 0; i < val.NumField(); i++ {
			if val.Type().Field(i).IsExported() {
				j.marshalGo(j.Sprintf("%s", val.Type().Field(i).Name), val.Field(i).Interface())
			}
		}
		j.out.WriteString("\n" + oldIndent + "}")
	}
}

type xdrGo struct {
	Value any
}

func (v xdrGo) XdrTypeName() string { return "go:" + reflect.ValueOf(v.Value).Type().String() }

func (v *xdrGo) XdrPointer() any { return &v.Value }

func (v xdrGo) XdrValue() any { return v.Value }

func (v *xdrGo) XdrMarshal(x xdr.XDR, name string) { x.Marshal(name, v) }

var (
	byteType = reflect.TypeOf(byte(0))
)

func AnyToJson(src any) (json []byte, err error) {
	switch v := src.(type) {
	case xdr.XdrType:
		return XdrToJson(v)
	default:
		return XdrToJson(&xdrGo{Value: src})
	}
}

// Format an XDR structure as JSON.
func XdrToJson(src xdr.XdrType) (json []byte, err error) {
	defer func() {
		if i := recover(); i != nil {
			json = nil
			err = i.(error)
		}
	}()
	j := &jsonOut{out: &bytes.Buffer{}}
	j.Marshal("", src)
	return j.out.Bytes(), nil
}

func MustXdrToJsonString(src xdr.XdrType) string {
	b, err := XdrToJson(src)
	must.NoError(context.TODO(), err)
	return string(b)
}

// utils

// Returns true for types whose field names should be hidden when the
// field name is of the form v[0-9]+.  This if for backwards
// compatibility, as fields of this type are intended to be used as
// versions of the same structure in a version union, and eliding the
// field names (v0, v1, ...) allows one to change the version without
// invalidating the rest of the fields.
func HideFieldName(field string, t xdr.XdrType) bool {
	if i := strings.LastIndexByte(field, '.'); i >= 0 {
		field = field[i+1:]
	}
	if !vField(field) {
		return false
	}
	return isDigitSandwich(t.XdrTypeName(), "TransactionV", "Envelope")
}

// Return true for field names of the form v[0-9]+
func vField(field string) bool {
	if len(field) < 2 || field[0] != 'v' {
		return false
	}
	for _, c := range field[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isDigitSandwich(target string, prefix string, suffix string) bool {
	if !strings.HasPrefix(target, prefix) {
		return false
	}
	target = target[len(prefix):]
	if len(target) <= len(suffix) || !strings.HasSuffix(target, suffix) {
		return false
	}
	target = target[:len(target)-len(suffix)]
	for _, r := range target {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
