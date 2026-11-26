package libxdr

import (
	"encoding/json"
	"fmt"
	"math"
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

// JSON numbers must decode exactly. A float64 detour prints 16777216 as
// 1.6777216e+07, which scans into a uint32 as 1.
func TestJsonToXdrDecodesLargeIntegersExactly(t *testing.T) {
	for _, counter := range []uint32{999_999, 1_000_000, 16_777_216, math.MaxUint32} {
		var got proto.SCPBallot
		src := fmt.Sprintf(`{"counter": %d, "value": ""}`, counter)
		if err := JsonToXdr(&got, []byte(src)); err != nil {
			t.Fatalf("counter %d: %v", counter, err)
		}
		if got.Counter != counter {
			t.Errorf("counter: got %d, want %d", got.Counter, counter)
		}
	}
	// Above 2^53 a float64 cannot hold every uint64.
	for _, nonce := range []uint64{1<<53 + 1, math.MaxUint64} {
		var got proto.PubkeyAccountPreimage
		src := fmt.Sprintf(`{"pubkey_hash": "%064x", "nonce": %d}`, 0, nonce)
		if err := JsonToXdr(&got, []byte(src)); err != nil {
			t.Fatalf("nonce %d: %v", nonce, err)
		}
		if got.Nonce != nonce {
			t.Errorf("nonce: got %d, want %d", got.Nonce, nonce)
		}
	}
	for _, height := range []int32{math.MinInt32, math.MaxInt32} {
		var got proto.BatonConfig
		src := fmt.Sprintf(`{"max_catchup_height": %d}`, height)
		if err := JsonToXdr(&got, []byte(src)); err != nil {
			t.Fatalf("height %d: %v", height, err)
		}
		if got.Max_catchup_height != height {
			t.Errorf("height: got %d, want %d", got.Max_catchup_height, height)
		}
	}
	for _, variable := range []int64{math.MinInt64, math.MaxInt64} {
		var got proto.HogAsk
		src := fmt.Sprintf(`{"variable": %d}`, variable)
		if err := JsonToXdr(&got, []byte(src)); err != nil {
			t.Fatalf("variable %d: %v", variable, err)
		}
		if got.Variable != variable {
			t.Errorf("variable: got %d, want %d", got.Variable, variable)
		}
	}
}

// An integer field takes only a whole integer token, never a prefix of one.
func TestJsonToXdrRejectsNonIntegerTokens(t *testing.T) {
	for _, value := range []string{`1e6`, `1.5`, `1.0`, `"12abc"`, `"12 34"`} {
		var got proto.SCPBallot
		src := fmt.Sprintf(`{"counter": %s, "value": ""}`, value)
		if err := JsonToXdr(&got, []byte(src)); err == nil {
			t.Errorf("counter %s: decoded as %d, want an error", value, got.Counter)
		}
	}
}

func TestJsonToXdrRejectsDataAfterTheValue(t *testing.T) {
	for _, src := range []string{`{"counter": 7, "value": ""} {}`, `{"counter": 7, "value": ""} x`} {
		var got proto.SCPBallot
		if err := JsonToXdr(&got, []byte(src)); err == nil {
			t.Errorf("%q: want an error", src)
		}
	}
}

// XdrToJson writes 32-bit integers as JSON numbers, so they must read back.
func TestJsonToXdrRoundTripsLargeUint32(t *testing.T) {
	want := proto.SCPBallot{Counter: 16_777_216, Value: proto.SCPValue{1}}
	s, err := XdrToJson(&want)
	if err != nil {
		t.Fatal(err)
	}
	var got proto.SCPBallot
	if err := JsonToXdr(&got, s); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// A union tag key is an integer token too.
func TestJsonToXdrRejectsNonIntegerUnionTags(t *testing.T) {
	var ok proto.NetworkConfig
	if err := JsonToXdr(&ok, []byte(`{"0": {"v0": {"params": []}}}`)); err != nil {
		t.Fatalf("tag 0: %v", err)
	}
	for _, tag := range []string{`0e3`, `0.5`, `0x`} {
		var got proto.NetworkConfig
		src := fmt.Sprintf(`{%q: {"v0": {"params": []}}}`, tag)
		if err := JsonToXdr(&got, []byte(src)); err == nil {
			t.Errorf("tag %s: decoded as %d, want an error", tag, got.V)
		}
	}
}
