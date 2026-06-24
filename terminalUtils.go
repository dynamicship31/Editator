package main

import "fmt"
import "strings"

var screenBuffer strings.Builder

// hscrollAnsi crops an already-colored (ANSI-escaped) string to a horizontal
// window, like hscroll does for plain text. Unlike hscroll, it understands
// escape sequences: it skips `offset` *visible* runes before it starts
// emitting output, but it always passes through every escape sequence it
// encounters along the way (even ones before the visible window) so that
// whatever color was active at the cut point is still active in the output.
// If width >= 0, output stops after `width` visible runes are emitted
// (escape sequences after that point are dropped too, since nothing visible
// follows them on this line).
func hscrollAnsi(s string, offset int, width int) string {
    runes := []rune(s)
    var out strings.Builder
    visibleSeen := 0
    visibleEmitted := 0
    stopped := false

    for i := 0; i < len(runes); i++ {
        if stopped { break }

        if runes[i] == '\x1b' {
            // Find the end of the escape sequence. We only generate CSI
            // sequences of the form ESC [ ... <final byte>, where the
            // final byte is in the range '@'..'~'. Copy the whole thing
            // verbatim, regardless of whether we're before, inside, or
            // after the visible window: color state needs to carry across
            // the crop point even though the bytes that set it are gone.
            j := i + 1
            if j < len(runes) && runes[j] == '[' {
                j++
                for j < len(runes) && !(runes[j] >= '@' && runes[j] <= '~') {
                    j++
                }
                if j < len(runes) { j++ } // include final byte
            }
            out.WriteString(string(runes[i:j]))
            i = j - 1
            continue
        }

        // Visible rune.
        if visibleSeen >= offset {
            if width >= 0 && visibleEmitted >= width {
                // Past the visible window on the right; stop entirely.
                // Any escape codes after this point only affect columns
                // that are off-screen, so there's nothing left to color.
                stopped = true
                continue
            }
            out.WriteRune(runes[i])
            visibleEmitted++
        }
        visibleSeen++
    }

    return out.String()
}

func printAt(x int, y int, msg string) {
    fmt.Fprintf(&screenBuffer, "\x1b[%d;%dH%s", y, x, msg)
}

func flushBuffer() {
    fmt.Print(screenBuffer.String())
    screenBuffer.Reset()
}

func cursorVisible(x bool) {
    if x { fmt.Fprintf(&screenBuffer, "\x1b[?25h") } else { fmt.Fprintf(&screenBuffer, "\x1b[?25l") }
}

func cursorAt(x int, y int) {
    fmt.Fprintf(&screenBuffer, "\x1b[%d;%dH", y, x)
}