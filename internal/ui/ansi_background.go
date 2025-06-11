package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getColorRGB converts a lipgloss.TerminalColor to 8-bit RGB values.
//
// # Parameters
//   - c: a lipgloss.TerminalColor
//
// # Returns
//   - r, g, b: 8-bit RGB values
func getColorRGB(c lipgloss.TerminalColor) (r8, g8, b8 uint8) {
	r, g, b, a := c.RGBA()
	// Un-premultiply alpha if needed
	if a > 0 && a < 0xffff {
		r = (r * 0xffff) / a
		g = (g * 0xffff) / a
		b = (b * 0xffff) / a
	}
	// Convert from 16-bit to 8-bit color
	r8 = uint8(r >> 8)
	g8 = uint8(g >> 8)
	b8 = uint8(b >> 8)
	return
}

var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

// ForceReplaceBackgroundWithLipgloss replaces any ANSI background color codes in 'input'
// with a single 24-bit background (48;2;R;G;B) using the provided newBgColor.
// This ensures consistent background theming for terminal UI output.
//
// # Usage
//
//	themed := ForceReplaceBackgroundWithLipgloss(someAnsiString, theme.Background)
//
// # Parameters
//   - input: the string containing ANSI escape sequences
//   - newBgColor: the desired background color (lipgloss.TerminalColor)
//
// # Returns
//   - The input string with all background color codes replaced by the new color
func ForceReplaceBackgroundWithLipgloss(input string, newBgColor lipgloss.TerminalColor) string {
	r, g, b := getColorRGB(newBgColor)
	newBg := fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	return ansiEscape.ReplaceAllStringFunc(input, func(seq string) string {
		const (
			escPrefixLen = 2 // "\x1b["
			escSuffixLen = 1 // "m"
		)
		raw := seq
		start := escPrefixLen
		end := len(raw) - escSuffixLen
		var sb strings.Builder
		sb.Grow((end - start) + len(newBg) + 2)
		i := start
		for i < end {
			j := i
			for j < end && raw[j] != ';' {
				j++
			}
			token := raw[i:j]
			if skip, nextIdx := shouldSkipBgToken(raw, i, j, end); skip {
				i = nextIdx
				continue
			}
			if keepToken(raw, i, j) {
				if sb.Len() > 0 {
					sb.WriteByte(';')
				}
				sb.WriteString(token)
			}
			i = j + 1
		}
		if sb.Len() > 0 {
			sb.WriteByte(';')
		}
		sb.WriteString(newBg)
		return "\x1b[" + sb.String() + "m"
	})
}

// shouldSkipBgToken determines if the current token is a background color code to skip.
func shouldSkipBgToken(raw string, i, j, end int) (skip bool, nextIdx int) {
	if len(raw[i:j]) == 2 && raw[i] == '4' && raw[i+1] == '8' {
		k := j + 1
		if k < end {
			l := k
			for l < end && raw[l] != ';' {
				l++
			}
			next := raw[k:l]
			if next == "5" {
				m := l + 1
				for m < end && raw[m] != ';' {
					m++
				}
				skip = true
				nextIdx = m + 1
				return
			} else if next == "2" {
				m := l + 1
				for count := 0; count < 3 && m < end; count++ {
					for m < end && raw[m] != ';' {
						m++
					}
					m++
				}
				skip = true
				nextIdx = m
				return
			}
		}
	}
	skip = false
	nextIdx = i
	return
}

// keepToken determines if a token should be kept in the output.
func keepToken(raw string, i, j int) bool {
	isNum := true
	val := 0
	for p := i; p < j; p++ {
		c := raw[p]
		if c < '0' || c > '9' {
			isNum = false
			break
		}
		val = val*10 + int(c-'0')
	}
	return !isNum || ((val < 40 || val > 47) && (val < 100 || val > 107) && val != 49)
}
