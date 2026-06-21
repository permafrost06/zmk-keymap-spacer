package main

import "testing"

func TestFixKeymapSpacing(t *testing.T) {
	input := "&kp TAB @mo FN\n&kp A &kp ENTER\n___ XXX &studio_unlock &bootloader &sys_reset\n&bt BT_SEL 1 &bt BT_CLR &kp A"
	want := "    &kp TAB        @mo FN         \n    &kp A          &kp ENTER      \n    ___            XXX            &studio_unlock &bootloader    &sys_reset     \n    &bt BT_SEL 1   &bt BT_CLR     &kp A          "

	got, err := fixKeymapSpacing(input, 15)
	if err != nil {
		t.Fatalf("fixKeymapSpacing returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapSpacingRejectsTooWideKeymap(t *testing.T) {
	_, err := fixKeymapSpacing("&kp ENTER", 5)
	if err == nil {
		t.Fatal("expected error")
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

	got, err := fixKeymapLayout(input, layout, 10, false)
	if err != nil {
		t.Fatalf("fixKeymapLayout returned error: %v", err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFixKeymapLayoutRejectsCountMismatch(t *testing.T) {
	_, err := fixKeymapLayout("&kp A &kp B", "x", 10, false)
	if err == nil {
		t.Fatal("expected error")
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
