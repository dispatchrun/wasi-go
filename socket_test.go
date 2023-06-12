package wasi_test

import (
	"reflect"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

func TestInet4AddressMarshalJSON(t *testing.T) {
	testMarshalJSON(t,
		&wasi.Inet4Address{
			Port: 4242,
			Addr: [4]byte{192, 168, 0, 2},
		},
		`"192.168.0.2:4242"`,
	)
}

func TestInet4AddressMarshalYAML(t *testing.T) {
	testMarshalYAML(t,
		&wasi.Inet4Address{
			Port: 4242,
			Addr: [4]byte{192, 168, 0, 2},
		},
		`192.168.0.2:4242`,
	)
}

func TestInet6AddressMarshalJSON(t *testing.T) {
	testMarshalJSON(t,
		&wasi.Inet6Address{
			Port: 4242,
			Addr: [16]byte{
				0x20, 0x01,
				0x0d, 0xb8,
				0x85, 0xa3,
				0x08, 0xd3,
				0x13, 0x19,
				0x8a, 0x2e,
				0x03, 0x70,
				0x73, 0x48,
			},
		},
		`"[2001:db8:85a3:8d3:1319:8a2e:370:7348]:4242"`,
	)
}

func TestInet6AddressMarshalYAML(t *testing.T) {
	testMarshalYAML(t,
		&wasi.Inet6Address{
			Port: 4242,
			Addr: [16]byte{
				0x20, 0x01,
				0x0d, 0xb8,
				0x85, 0xa3,
				0x08, 0xd3,
				0x13, 0x19,
				0x8a, 0x2e,
				0x03, 0x70,
				0x73, 0x48,
			},
		},
		`[2001:db8:85a3:8d3:1319:8a2e:370:7348]:4242`,
	)
}

func testMarshalJSON(t *testing.T, addr wasi.SocketAddress, want string) {
	b, err := addr.(interface{ MarshalJSON() ([]byte, error) }).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != want {
		t.Errorf("%q != %q", b, want)
	}
}

func testMarshalYAML(t *testing.T, addr wasi.SocketAddress, want any) {
	v, err := addr.(interface{ MarshalYAML() (any, error) }).MarshalYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, want) {
		t.Errorf("%#v != %#v", v, want)
	}
}
