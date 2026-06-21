# zmk-keymap-spacer

Formats ZMK-style keymaps in-place from Vim or Neovim using `!` bang filters. It can align keymaps to a fixed width, place them into a visual `.layout` mask, and generate border comments that can be safely reformatted again.

## Install

Install with Go:

```sh
go install .
```

This installs `zmk-keymap-spacer` into your Go binary directory, usually `~/go/bin`.

Make sure that directory is in your `PATH`:

```sh
export PATH="$HOME/go/bin:$PATH"
```

You can also use the Makefile:

```sh
make install
```

Then call the tool from Vim or Neovim without `./`:

```vim
:'<,'>!zmk-keymap-spacer -width 13 -layout sofleez.layout -split-middle
```

## Build Locally

```sh
go build -o zmk-keymap-spacer .
```

If you build locally instead of installing, call it by path from your keymap file, for example `./zmk-keymap-spacer`.

## Vim/Neovim Usage

The primary workflow is formatting a selected block inside a ZMK keymap file.

Select the keymap lines in visual mode, then run:

```vim
:'<,'>!./zmk-keymap-spacer -width 13 -layout sofleez.layout
```

For a split keyboard layout:

```vim
:'<,'>!./zmk-keymap-spacer -width 13 -layout sofleez.layout -split-middle
```

For simple alignment without a layout:

```vim
:'<,'>!./zmk-keymap-spacer -width 13
```

Format the current paragraph without a visual selection:

```vim
vip:!./zmk-keymap-spacer -width 13
```

Generated border lines start with `//`, and input lines starting with `//` are ignored. That means you can run the same bang command repeatedly over already formatted output.

Example repeated formatting command:

```vim
:'<,'>!./zmk-keymap-spacer -width 13 -layout sofleez.layout -split-middle
```

## Plain Formatting

Format keymaps from stdin with each keymap taking `13` characters:

```sh
printf '&kp TAB &mo FN\n&kp A &kp ENTER\n' | ./zmk-keymap-spacer -width 13
```

Output:

```text
    &kp TAB      &mo FN
    &kp A        &kp ENTER
```

Each output line starts with four spaces.

## Layout Formatting

A `.layout` file uses `x` for keymap slots. Spaces are preserved as gaps.

`sofleez.layout`:

```text
xxxxxx  xxxxxx
xxxxxx  xxxxxx
xxxxxx  xxxxxx
xxxxxxxxxxxxxx
  xxxx  xxxx
```

Run the formatter:

```sh
./zmk-keymap-spacer -width 13 -layout sofleez.layout < keymap.txt
```

Example input:

```text
&kp GRAVE &kp N1 &kp N2 &kp N3 &kp N4 &kp N5 &kp N6 &kp N7 &kp N8 &kp N9 &kp N0 &kp MINUS
&kp TAB &kp Q &kp W &kp E &kp R &kp T &kp Y &kp U &kp I &kp O &kp P &kp BSLH
&kp ESC &kp A &kp S &kp D &kp F &kp G &kp H &kp J &kp K &kp L &kp SEMI &kp SQT
&kp FN &kp Z &kp X &kp C &kp V &kp B &kp C_MUTE &kp F13 &kp N &kp M &kp COMMA &kp DOT &kp FSLH &kp RALT
&kp LSHFT &mo SYS &kp LCTRL &kp SPACE &kp RET &mo FN &kp BSPC &kp RGUI
```

Output includes border comments:

```text
//╭────────────┬────────────┬────────────┬────────────┬────────────┬────────────╮                          ╭────────────┬────────────┬────────────┬────────────┬────────────┬────────────╮
    &kp GRAVE    &kp N1       &kp N2       &kp N3       &kp N4       &kp N5                                 &kp N6       &kp N7       &kp N8       &kp N9       &kp N0       &kp MINUS
```

## Split Middle

For split keyboards, add `-split-middle` to visually separate a continuous middle row:

```sh
./zmk-keymap-spacer -width 13 -layout sofleez.layout -split-middle < keymap.txt
```

This turns a continuous row like `xxxxxxxxxxxxxx` into a split row in the generated borders:

```text
//├────────────┼────────────┼────────────┼────────────┼────────────┼────────────┼────────────╮ ╭────────────┼────────────┼────────────┼────────────┼────────────┼────────────┼────────────┤
```

### Small Layout Examples

`mini.layout`:

```text
xx  xx
xxxxxx
```

Input:

```text
&kp Q &kp W &kp E &kp R
&kp A &kp S &kp D &kp F &kp G &kp H
```

Without `-split-middle`:

```sh
./zmk-keymap-spacer -width 8 -layout mini.layout < keymap.txt
```

Output:

```text
//╭───────┬───────╮                ╭───────┬───────╮
    &kp Q   &kp W                   &kp E   &kp R
//├───────┼───────┼───────┬───────┼───────┼───────┤
    &kp A   &kp S   &kp D   &kp F   &kp G   &kp H
//╰───────┴───────┴───────┴───────┴───────┴───────╯
```

With `-split-middle`:

```sh
./zmk-keymap-spacer -width 8 -layout mini.layout -split-middle < keymap.txt
```

Output:

```text
//╭───────┬───────╮                 ╭───────┬───────╮
    &kp Q   &kp W                    &kp E   &kp R
//├───────┼───────┼───────╮ ╭───────┼───────┼───────┤
    &kp A   &kp S   &kp D     &kp F   &kp G   &kp H
//╰───────┴───────┴───────╯ ╰───────┴───────┴───────╯
```

## Ignored Lines

Input lines starting with `//` are ignored. This lets you re-run the formatter over already formatted output from Vim or Neovim:

```sh
:'<,'>!./zmk-keymap-spacer -width 13 -layout sofleez.layout -split-middle
```

Indented comment lines are ignored too.

## Supported Keymaps

Two-part keymaps:

```text
&kp TAB
@mo FN
&bt BT_CLR
```

Three-part keymaps:

```text
&bt BT_SEL 1
```

One-part keymaps:

```text
___
XXX
&studio_unlock
&bootloader
&sys_reset
```

If a keymap is longer than `-width`, the program exits with an error.
