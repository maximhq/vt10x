package vt10x

import (
	"strconv"
	"testing"
)

func TestSTRParse(t *testing.T) {
	var str strEscape
	str.reset()
	str.buf = []rune("0;some text")
	str.parse()
	if str.arg(0, 17) != 0 || str.argString(1, "") != "some text" {
		t.Fatal("STR parse mismatch")
	}
}

func TestParseColor(t *testing.T) {
	type testCase struct {
		name    string
		input   string
		r, g, b int
	}

	for _, tc := range []testCase{
		{
			"rgb 4 bit zero",
			"rgb:0/0/0",
			0, 0, 0,
		},
		{
			"rgb 4 bit max",
			"rgb:f/f/f",
			255, 255, 255,
		},
		{
			"rgb 4 bit values",
			"rgb:1/2/3",
			17, 34, 51,
		},
		{
			"rgb 8 bit zero",
			"rgb:00/00/00",
			0, 0, 0,
		},
		{
			"rgb 8 bit max",
			"rgb:ff/ff/ff",
			255, 255, 255,
		},
		{
			"rgb 8 bit values",
			"rgb:11/22/33",
			17, 34, 51,
		},
		{
			"rgb 12 bit zero",
			"rgb:000/000/000",
			0, 0, 0,
		},
		{
			"rgb 12 bit max",
			"rgb:fff/fff/fff",
			255, 255, 255,
		},
		{
			"rgb 12 bit values",
			"rgb:111/222/333",
			17, 34, 51,
		},
		{
			"rgb 16 bit zero",
			"rgb:0000/0000/0000",
			0, 0, 0,
		},
		{
			"rgb 16 bit max",
			"rgb:ffff/ffff/ffff",
			255, 255, 255,
		},
		{
			"rgb 16 bit values",
			"rgb:1111/2222/3333",
			17, 34, 51,
		},
		{
			"rgb 16 bit values",
			"rgb:1111/2222/3333",
			17, 34, 51,
		},
		{
			"hash 4 bit zero",
			"#000",
			0, 0, 0,
		},
		{
			"hash 4 bit max",
			"#fff",
			240, 240, 240,
		},
		{
			"hash 4 bit values",
			"#123",
			16, 32, 48,
		},
		{
			"hash 8 bit zero",
			"#000000",
			0, 0, 0,
		},
		{
			"hash 8 bit max",
			"#ffffff",
			255, 255, 255,
		},
		{
			"hash 8 bit values",
			"#112233",
			17, 34, 51,
		},
		{
			"hash 12 bit zero",
			"#000000000",
			0, 0, 0,
		},
		{
			"hash 12 bit max",
			"#fffffffff",
			255, 255, 255,
		},
		{
			"hash 12 bit values",
			"#111222333",
			17, 34, 51,
		},
		{
			"hash 16 bit zero",
			"#000000000000",
			0, 0, 0,
		},
		{
			"hash 16 bit max",
			"#ffffffffffff",
			255, 255, 255,
		},
		{
			"hash 16 bit values",
			"#111122223333",
			17, 34, 51,
		},
		{
			"rgb upper case",
			"RGB:0/A/F",
			0, 170, 255,
		},
		{
			"hash upper case",
			"#FFF",
			240, 240, 240,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b, err := parseColor(tc.input)
			if err != nil {
				t.Fatalf("failed to parse color: %s", err)
			}

			if r != tc.r || g != tc.g || b != tc.b {
				t.Fatalf("expected (%d, %d, %d), got (%d, %d, %d)", tc.r, tc.g, tc.b, r, g, b)
			}
		})
	}
}

func TestSetColorNameBounds(t *testing.T) {
	spec := "rgb:11/22/33"

	type testCase struct {
		name    string
		j       int
		wantErr bool
	}

	cases := []testCase{
		{"palette zero", 0, false},
		{"palette max", 1<<24 - 1, false},
		{"default fg", int(DefaultFG), false},
		{"default bg", int(DefaultBG), false},
		{"default cursor", int(DefaultCursor), false},
		{"negative", -1, true},
		{"above max", 1<<24 + 3, true},
	}

	// Color is a uint32, so on platforms with a 64-bit int this would silently
	// truncate to palette entry 0 if converted without a bounds check.
	if strconv.IntSize > 32 {
		// Computed via a variable so the conversion stays non-constant and the
		// test still compiles where int is 32 bits.
		overflow := int64(1<<32-1) + 1 // math.MaxUint32 + 1
		cases = append(cases, testCase{"truncates to palette zero", int(overflow), true})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &State{colorOverride: make(map[Color]Color)}
			err := state.setColorName(tc.j, &spec)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("setColorName(%d) = nil, want error", tc.j)
				}
				if len(state.colorOverride) != 0 {
					t.Fatalf("setColorName(%d) wrote %v, want no override", tc.j, state.colorOverride)
				}
				return
			}
			if err != nil {
				t.Fatalf("setColorName(%d) = %v, want nil", tc.j, err)
			}
			if got, ok := state.colorOverride[Color(tc.j)]; !ok || got != Color(0x112233) {
				t.Fatalf("colorOverride[%d] = %v, %v; want 0x112233, true", tc.j, got, ok)
			}
		})
	}
}
