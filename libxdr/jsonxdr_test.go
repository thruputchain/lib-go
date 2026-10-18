package libxdr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	proto "github.com/thruputchain/thruput-protocol-go"
)

func TestUnion(t *testing.T) {
	x := proto.SCPStatement{}
	x.Pledges.Stmt_type = proto.SCP_ST_CONFIRM
	*x.Pledges.Confirm() = proto.SCPStatementConfirm{
		Ballot:          proto.SCPBallot{Counter: 5, Value: proto.SCPValue{1, 2, 3}},
		N_prepared:      1,
		N_commit:        2,
		N_h:             3,
		Quorum_set_hash: proto.SCPHash{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 4}, // SCPHash
	}
	s, err := XdrToJson(&x)
	if err != nil {
		t.Error(err)
	}
	println(string(s))
	y := proto.SCPStatement{}
	err = JsonToXdr(&y, s)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(x, y) {
		t.Errorf("expecting equality")
	}
}

func TestUnionTaggedJsonShape(t *testing.T) {
	pk := proto.PublicKey{Scheme: proto.Ed25519}
	*pk.Ed25519() = proto.Ed25519PublicKey{0x2e, 0xa0, 0xf2, 0x40}

	s, err := XdrToJson(&pk)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(s, &got); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"Ed25519": map[string]any{
			"ed25519": "2ea0f24000000000000000000000000000000000000000000000000000000000",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected union json shape: got=%v want=%v", got, want)
	}
}

func TestUnionAcceptsLegacyAndTaggedForms(t *testing.T) {
	tagged := []byte(`{"Ed25519":{"ed25519":"2ea0f24000000000000000000000000000000000000000000000000000000000"}}`)
	legacy := []byte(`{"scheme":"Ed25519","ed25519":"2ea0f24000000000000000000000000000000000000000000000000000000000"}`)

	var a proto.PublicKey
	if err := JsonToXdr(&a, tagged); err != nil {
		t.Fatal(err)
	}
	if a.Scheme != proto.Ed25519 || a.Ed25519() == nil || (*a.Ed25519())[0] != 0x2e {
		t.Fatalf("tagged form did not unmarshal as expected: %+v", a)
	}

	var b proto.PublicKey
	if err := JsonToXdr(&b, legacy); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("tagged and legacy forms should decode equally: tagged=%+v legacy=%+v", a, b)
	}
}

func TestRoundtrip(t *testing.T) {
	x := proto.ValidatorPublicInfo{Address: "1.2.3.4"}
	s, err := XdrToJson(&x)
	if err != nil {
		t.Error(err)
	}
	y := proto.ValidatorPublicInfo{}
	err = JsonToXdr(&y, s)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(x, y) {
		t.Errorf("expecting equality")
	}
}

type testNonXdrType struct {
	Bool           bool
	Int32          int32
	Int64          int64
	Unexported     bool
	ByteArray      [3]byte
	SomeXdr        proto.SHA256Hash
	Complex        complex128
	MissingPointer *[]string
	PresentPointer *string
	SliceString    []string
	Bytes          []byte
	EmbeddedXDR    proto.ValidatorPublicInfo
}

func TestAnyToJson(t *testing.T) {
	ss := "prezo"
	v := &testNonXdrType{
		SliceString:    []string{"abc", "def"},
		Bytes:          []byte{0, 1, 2},
		PresentPointer: &ss,
		EmbeddedXDR:    proto.ValidatorPublicInfo{Address: "1.2.3.4"},
	}
	s, err := AnyToJson(v)
	if err != nil {
		t.Error(err)
	}
	expected := `
{
    "Bool": false,
    "Int32": 0,
    "Int64": "0",
    "Unexported": false,
    "ByteArray": "000000",
    "SomeXdr": "0000000000000000000000000000000000000000000000000000000000000000",
    "Complex": (0+0i),
    "MissingPointer": <nil>,
    "PresentPointer": "prezo",
    "SliceString": [
        "abc",
        "def"
    ],
    "Bytes": "000102",
    "EmbeddedXDR": {
        "Public_key": {
            "Scheme": Ed25519
        },
        "Address": "1.2.3.4"
    }
}`
	if string(s) != expected {
		t.Errorf("expecting %s, got %s", expected, string(s))
	}
	fmt.Println(string(s))
}
