package main

import (
	"strings"
	"testing"
)

func TestFixKeymapSpacing(t *testing.T) {
	input := "&kp TAB @mo FN\n&kp A &kp ENTER\n___ XXX &studio_unlock &bootloader &sys_reset\n&bt BT_SEL 1 &bt BT_CLR &kp A"
	want := "    &kp TAB        @mo FN         \n    &kp A          &kp ENTER      \n    ___            XXX            &studio_unlock &bootloader    &sys_reset     \n    &bt BT_SEL 1   &bt BT_CLR     &kp A          "

	got, err := fixKeymapSpacing(input, 15, "    ")
	if err != nil {
		t.Fatalf("fixKeymapSpacing returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapSpacingRejectsTooWideKeymap(t *testing.T) {
	_, err := fixKeymapSpacing("&kp ENTER", 5, "    ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFixKeymapSpacingUsesLongestKeymapWhenWidthOmitted(t *testing.T) {
	got, err := fixKeymapSpacing("&kp A &kp ENTER", 0, "    ")
	if err != nil {
		t.Fatalf("fixKeymapSpacing returned error: %v", err)
	}
	want := "    &kp A       &kp ENTER   "
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapSpacingUsesCustomIndent(t *testing.T) {
	got, err := fixKeymapSpacing("&kp A", 10, "\t")
	if err != nil {
		t.Fatalf("fixKeymapSpacing returned error: %v", err)
	}
	want := "\t&kp A     "
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveWidth(t *testing.T) {
	tests := []struct {
		name  string
		spec  string
		input string
		want  int
	}{
		{name: "exact", spec: "13", input: "&kp A", want: 13},
		{name: "relative", spec: "+2", input: "&kp A &kp ENTER", want: 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveWidth(tt.input, tt.spec)
			if err != nil {
				t.Fatalf("resolveWidth returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFixKeymapLayout(t *testing.T) {
	input := "&kp A &kp B &kp C &kp D"
	layout := "xx\nxx"
	top := "//╭─────────┬─────────╮\n"
	middle := "//├─────────┼─────────┤\n"
	bottom := "//╰─────────┴─────────╯"
	want := top +
		"    &kp A     &kp B\n" + middle +
		"    &kp C     &kp D\n" + bottom

	got, err := fixKeymapLayout(input, layout, 10, false, "    ")
	if err != nil {
		t.Fatalf("fixKeymapLayout returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapLayoutRejectsCountMismatch(t *testing.T) {
	_, err := fixKeymapLayout("&kp A &kp B", "x", 10, false, "    ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFixKeymapLayoutUsesLongestKeymapWhenWidthOmitted(t *testing.T) {
	got, err := fixKeymapLayout("&kp A &kp ENTER", "xx", 0, false, "    ")
	if err != nil {
		t.Fatalf("fixKeymapLayout returned error: %v", err)
	}
	want := "//╭───────────┬───────────╮\n    &kp A       &kp ENTER\n//╰───────────┴───────────╯"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapLayoutCentersFiveSlotLabel(t *testing.T) {
	got, err := fixKeymapLayout("_BT_SEL_KEYS_ &kp A", "xxxxxx", 10, false, "    ")
	if err != nil {
		t.Fatalf("fixKeymapLayout returned error: %v", err)
	}
	want := "//╭─────────┬─────────┬─────────┬─────────┬─────────┬─────────╮\n" +
		"    " + strings.Repeat(" ", 18) + "_BT_SEL_KEYS_" + strings.Repeat(" ", 19) + "&kp A\n" +
		"//╰─────────┴─────────┴─────────┴─────────┴─────────┴─────────╯"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestParseKeymapsIgnoresCommentLines(t *testing.T) {
	got, err := parseKeymaps("// border\n&kp A\n    // another border\n&kp B", 10)
	if err != nil {
		t.Fatalf("parseKeymaps returned error: %v", err)
	}
	want := []string{"&kp A", "&kp B"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestParseKeymapsSupportsAdditionalForms(t *testing.T) {
	got, err := parseKeymaps("&hml LSHFT D &hmr RSHFT K &caps_word _BT_SEL_KEYS_", 0)
	if err != nil {
		t.Fatalf("parseKeymaps returned error: %v", err)
	}
	want := []string{"&hml LSHFT D", "&hmr RSHFT K", "&caps_word", "_BT_SEL_KEYS_"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
