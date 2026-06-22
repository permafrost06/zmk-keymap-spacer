package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

var onePartKeymaps = []string{"___", "XXX", "_BT_SEL_KEYS_", "&studio_unlock", "&bootloader", "&sys_reset", "&caps_word"}

func resolveWidth(input string, spec string) (int, error) {
	if strings.HasPrefix(spec, "+") {
		extra, err := strconv.Atoi(strings.TrimPrefix(spec, "+"))
		if err != nil || extra < 0 {
			return 0, fmt.Errorf("-width relative value must look like +3")
		}

		width, err := longestKeymapLength(input)
		if err != nil {
			return 0, err
		}
		return width + extra, nil
	}

	width, err := strconv.Atoi(spec)
	if err != nil || width <= 0 {
		return 0, fmt.Errorf("-width must be an integer greater than 0 or a relative value like +3")
	}
	return width, nil
}

func main() {
	widthSpec := flag.String("width", "+3", "keymap width: exact integer like 13, or relative value like +3")
	indent := flag.String("indent", "    ", "string written at the start of each keymap output line")
	input := flag.String("input", "", "multiline keymap string; stdin is used when omitted")
	layoutPath := flag.String("layout", "", "path to a .layout file; each x is filled with one keymap")
	splitMiddle := flag.Bool("split-middle", false, "split continuous middle rows into left and right halves")
	flag.Parse()

	text := *input
	if text == "" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
			os.Exit(1)
		}
		text = string(data)
	}

	width, err := resolveWidth(text, *widthSpec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var fixed string
	if *layoutPath != "" {
		layout, readErr := os.ReadFile(*layoutPath)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "read layout: %v\n", readErr)
			os.Exit(1)
		}

		fixed, err = fixKeymapLayout(text, string(layout), width, *splitMiddle, *indent)
	} else {
		fixed, err = fixKeymapSpacing(text, width, *indent)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(fixed)
}

func fixKeymapSpacing(input string, width int, indent string) (string, error) {
	if width == 0 {
		var err error
		width, err = longestKeymapLength(input)
		if err != nil {
			return "", err
		}
		width += 3
	}

	var out strings.Builder

	for lineStart := 0; lineStart < len(input); {
		lineEnd := lineStart
		for lineEnd < len(input) && input[lineEnd] != '\n' && input[lineEnd] != '\r' {
			lineEnd++
		}

		fixed, err := fixLineSpacing(input[lineStart:lineEnd], width, indent)
		if err != nil {
			return "", err
		}
		out.WriteString(fixed)

		if lineEnd == len(input) {
			break
		}
		out.WriteByte(input[lineEnd])
		lineStart = lineEnd + 1
		if input[lineEnd] == '\r' && lineStart < len(input) && input[lineStart] == '\n' {
			out.WriteByte(input[lineStart])
			lineStart++
		}
	}

	return out.String(), nil
}

func fixKeymapLayout(input string, layout string, width int, splitMiddle bool, indent string) (string, error) {
	if width == 0 {
		var err error
		width, err = longestKeymapLength(input)
		if err != nil {
			return "", err
		}
		width += 3
	}

	keymaps, err := parseKeymaps(input, width)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	keymapIndex := 0
	rows := strings.Split(strings.TrimRight(layout, "\r\n"), "\n")
	for rowIndex, row := range rows {
		border := layoutTopLine(row, width, splitMiddle)
		if rowIndex > 0 {
			border = layoutTransitionLine(rows[rowIndex-1], row, width, splitMiddle)
		}
		out.WriteString("//")
		out.WriteString(border)
		out.WriteByte('\n')

		out.WriteString(indent)
		var rowOut strings.Builder
		splitIndex := splitMiddleIndex(row, splitMiddle)
		for i := 0; i < len(row); i++ {
			r := rune(row[i])
			if r == ' ' {
				writeKeyRowGap(&rowOut, row, i, width, splitMiddle)
				continue
			}
			if r != 'x' {
				rowOut.WriteRune(r)
				continue
			}
			if keymapIndex == len(keymaps) {
				return "", fmt.Errorf("layout has more slots than keymaps")
			}

			keymap := keymaps[keymapIndex]
			span := keymapSpan(keymap)
			if span > 1 {
				if countSlots(row[i:]) < span {
					return "", fmt.Errorf("keymap %q needs %d layout slots", keymap, span)
				}

				rowOut.WriteString(center(keymap, width*span))
				keymapIndex++
				i += span - 1
				continue
			}

			rowOut.WriteString(keymap)
			rowOut.WriteString(strings.Repeat(" ", width-len(keymap)))
			keymapIndex++
			if splitIndex == i+1 {
				rowOut.WriteString("  ")
			}
		}

		out.WriteString(strings.TrimRight(rowOut.String(), " "))
		out.WriteByte('\n')
	}
	if len(rows) > 0 {
		out.WriteString("//")
		out.WriteString(layoutBottomLine(rows[len(rows)-1], width, splitMiddle))
	}

	if keymapIndex < len(keymaps) {
		return "", fmt.Errorf("layout has %d slots but input has %d keymaps", keymapIndex, len(keymaps))
	}

	return out.String(), nil
}

func layoutTopLine(row string, width int, splitMiddle bool) string {
	return layoutSimpleBorderLine(row, width, splitMiddle, "╭", "┬", "╮")
}

func layoutBottomLine(row string, width int, splitMiddle bool) string {
	return layoutSimpleBorderLine(row, width, splitMiddle, "╰", "┴", "╯")
}

func layoutSimpleBorderLine(row string, width int, splitMiddle bool, left string, join string, right string) string {
	var out strings.Builder
	dashes := strings.Repeat("─", width-1)
	row = splitMiddleRow(row, splitMiddle)
	for i := 0; i < len(row); {
		if row[i] == ' ' {
			writeLayoutGap(&out, row, i, width, splitMiddle)
			i++
			continue
		}
		if row[i] == '|' {
			out.WriteByte(' ')
			i++
			continue
		}
		if row[i] != 'x' {
			out.WriteByte(row[i])
			i++
			continue
		}

		runStart := i
		for i < len(row) && row[i] == 'x' {
			i++
		}
		out.WriteString(left)
		for cell := runStart; cell < i; cell++ {
			out.WriteString(dashes)
			if cell == i-1 {
				out.WriteString(right)
			} else {
				out.WriteString(join)
			}
		}
	}

	return out.String()
}

func layoutTransitionLine(prev string, current string, width int, splitMiddle bool) string {
	prev = splitMiddleRow(prev, splitMiddle)
	current = splitMiddleRow(current, splitMiddle)
	maxLen := max(len(prev), len(current))
	firstCurrent := -1
	lastCurrent := -1
	for i := 0; i < maxLen; i++ {
		if charAt(current, i) == 'x' {
			if firstCurrent == -1 {
				firstCurrent = i
			}
			lastCurrent = i
		}
	}

	var out strings.Builder
	dashes := strings.Repeat("─", width-1)
	for i := 0; i < maxLen; i++ {
		if charAt(prev, i) == '|' || charAt(current, i) == '|' {
			out.WriteByte(' ')
			continue
		}
		p := charAt(prev, i) == 'x'
		c := charAt(current, i) == 'x'
		if i >= firstCurrent && i <= lastCurrent && (p || c) {
			p = true
			c = true
		}
		if !p && !c {
			writeLayoutGap(&out, current, i, width, splitMiddle)
			continue
		}

		startUp := p
		startDown := c
		if charAt(prev, i-1) == '|' || charAt(current, i-1) == '|' {
			currentCell := charAt(current, i-1) == '|'
			startUp = !currentCell
			startDown = currentCell
		}
		out.WriteString(borderRune(charAt(prev, i-1) == 'x' || charAt(current, i-1) == 'x', true, startUp, startDown))
		out.WriteString(dashes)
		if i == maxLen-1 || (!p && !c) || (!cellAt(prev, current, i+1)) {
			endUp := p
			endDown := c
			if charAt(prev, i+1) == '|' || charAt(current, i+1) == '|' {
				currentCell := charAt(current, i+1) == '|'
				endUp = !currentCell
				endDown = currentCell
			}
			out.WriteString(borderRune(true, false, endUp, endDown))
		}
	}
	return out.String()
}

func writeLayoutGap(out *strings.Builder, row string, i int, width int, splitMiddle bool) {
	out.WriteString(strings.Repeat(" ", width))
	if splitMiddle && i > 0 && i < len(row)-1 && row[i-1] == ' ' && strings.Contains(row[:i], "x") && strings.Contains(row[i+1:], "x") {
		out.WriteByte(' ')
	}
}

func writeKeyRowGap(out *strings.Builder, row string, i int, width int, splitMiddle bool) {
	out.WriteString(strings.Repeat(" ", width))
	if splitMiddle && strings.Contains(row[:i], "x") && strings.Contains(row[i+1:], "x") {
		out.WriteByte(' ')
	}
}

func keymapSpan(keymap string) int {
	if keymap == "_BT_SEL_KEYS_" {
		return 5
	}
	return 1
}

func countSlots(row string) int {
	count := 0
	for i := 0; i < len(row); i++ {
		if row[i] == 'x' {
			count++
			continue
		}
		if row[i] != '|' {
			break
		}
	}
	return count
}

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func borderRune(left bool, right bool, up bool, down bool) string {
	switch {
	case left && right && up && down:
		return "┼"
	case left && right && up:
		return "┴"
	case left && right && down:
		return "┬"
	case right && up && down:
		return "├"
	case left && up && down:
		return "┤"
	case right && down:
		return "╭"
	case left && down:
		return "╮"
	case right && up:
		return "╰"
	case left && up:
		return "╯"
	default:
		return " "
	}
}

func cellAt(prev string, current string, i int) bool {
	return charAt(prev, i) == 'x' || charAt(current, i) == 'x'
}

func charAt(s string, i int) byte {
	if i < 0 || i >= len(s) {
		return ' '
	}
	return s[i]
}

func splitMiddleRow(row string, splitMiddle bool) string {
	middle := splitMiddleIndex(row, splitMiddle)
	if middle == -1 {
		return row
	}
	return row[:middle] + "|" + row[middle:]
}

func splitMiddleIndex(row string, splitMiddle bool) int {
	if !splitMiddle || strings.Contains(row, " ") || strings.Count(row, "x")%2 != 0 {
		return -1
	}
	return len(row) / 2
}

func parseKeymaps(input string, width int) ([]string, error) {
	var cleaned strings.Builder
	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		cleaned.WriteString(line)
		cleaned.WriteByte('\n')
	}

	fields := strings.Fields(cleaned.String())
	keymaps := make([]string, 0, len(fields))
	for i := 0; i < len(fields); {
		keymapLen := keymapFieldCount(fields[i:])
		if keymapLen == 0 {
			return nil, fmt.Errorf("unexpected token %q", fields[i])
		}

		keymap := strings.Join(fields[i:i+keymapLen], " ")
		if width > 0 && len(keymap) > width*keymapSpan(keymap) {
			return nil, fmt.Errorf("keymap %q is longer than width %d", keymap, width)
		}
		keymaps = append(keymaps, keymap)
		i += keymapLen
	}

	return keymaps, nil
}

func longestKeymapLength(input string) (int, error) {
	keymaps, err := parseKeymaps(input, 0)
	if err != nil {
		return 0, err
	}
	width := 0
	for _, keymap := range keymaps {
		width = max(width, len(keymap))
	}
	if width == 0 {
		return 0, fmt.Errorf("no keymaps found")
	}
	return width, nil
}

func fixLineSpacing(line string, width int, indent string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return indent, nil
	}

	var out strings.Builder
	out.WriteString(indent)
	for i := 0; i < len(fields); {
		keymapLen := keymapFieldCount(fields[i:])
		if keymapLen == 0 {
			if out.Len() > 0 {
				out.WriteByte(' ')
			}
			out.WriteString(fields[i])
			i++
			continue
		}

		keymap := strings.Join(fields[i:i+keymapLen], " ")
		if len(keymap) > width {
			return "", fmt.Errorf("keymap %q is longer than width %d", keymap, width)
		}

		out.WriteString(keymap)
		out.WriteString(strings.Repeat(" ", width-len(keymap)))
		i += keymapLen
	}

	return out.String(), nil
}

func keymapFieldCount(fields []string) int {
	if len(fields) == 0 {
		return 0
	}
	if slices.Contains(onePartKeymaps, fields[0]) {
		return 1
	}
	if fields[0] == "&bt" && len(fields) >= 2 && fields[1] == "BT_CLR" {
		return 2
	}
	if fields[0] == "&bt" && len(fields) >= 3 && fields[1] == "BT_SEL" {
		return 3
	}
	if (fields[0] == "&hml" || fields[0] == "&hmr") && len(fields) >= 3 {
		return 3
	}
	if (strings.HasPrefix(fields[0], "&") || strings.HasPrefix(fields[0], "@")) && len(fields) >= 2 {
		return 2
	}
	return 0
}
