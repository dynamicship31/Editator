package main

// I'm prob gonna stop at alpha 4 for now (0.4) since I lowk don't want to spend my whole life making an editor, but who knows, maybe one day I'll drop 0.5 or something.

// If I drop the next update it will have copy and paste support.

import (
    "fmt"
    "time"
    "strings"
    "strconv"
    "github.com/eiannone/keyboard"
    "golang.org/x/term"
    "os"
    "os/exec"
    "sync"
)

type KeyState struct {
    Char rune
    Key  keyboard.Key
}

// Kind values for UndoAction.
const undoEdit   int8 = 0 // A line's text changed in place (old_str -> new_str).
const undoInsert int8 = 1 // A brand new line was inserted at line_num (undo must delete it).
const undoDelete int8 = 2 // A line was removed from line_num (undo must re-insert old_str there).

type UndoAction struct {
    kind     int8
    line_num int
    old_str  string
    new_str  string
}

// UndoGroup is a set of UndoActions.
type UndoGroup struct {
    actions    []UndoAction
    cursor_pos [2]int // cursor position BEFORE the edit, restored on undo
}

const normalMode       int8 = 0
const commandMode      int8 = 1
const wordMode         int8 = 2
const pluginMode       int8 = 3 // COMING SOON (maybe)
const visualMode       int8 = 4 // WIP

var cursor_pos         [2]int
var show_pos           [2]int
var sticky_column      int = 1
var scroll             int = 0
var scroll_x           int = 0
var globalKey          KeyState
var loadedFile         []string
var nameFile           string
var fileType           string
var mode               int8
var globalAsk          string
var globalPrompt       string

var cursor_line_count_bg      string = "\x1b[48;2;65;65;65m"
var cursor_line_count_fg      string = "\x1b[1;33m"
var line_count_bg             string = "\x1b[48;2;34;34;34m"
var line_count_fg             string = ""
var bg_color                  string = "\x1b[48;2;31;31;31m"
var bottom_bar_bg             string = "\x1b[48;2;24;24;24m"
var bottom_bar_fg             string = ""
var cursor_type               string = "\x1b[2 q"

var auto_indenting      bool   = true
var highlighting        bool   = true
var toggle_mode_after   bool   = false

var undo_stack          = []UndoGroup{} // History of applied edits, most recent last.
var redo_stack          = []UndoGroup{} // Groups popped off undo_stack by undo(), most recent last.
const undo_max_size     int = 1000      // Cap so memory doesn't grow forever on long sessions.

var mu sync.Mutex

func insertLine(index int) { loadedFile = append(loadedFile[:index], append([]string{""}, loadedFile[index:]...)...) }

func insertAtLine(lineIndex int, charIndex int, value string) {
    runes := []rune(loadedFile[lineIndex])
    loadedFile[lineIndex] = string(runes[:charIndex]) + value + string(runes[charIndex:])
}

func undo() {
    if len(undo_stack) == 0 { return }

    group := undo_stack[len(undo_stack)-1]
    undo_stack = undo_stack[:len(undo_stack)-1]

    // Apply the group's actions in reverse order so multiline groups.
    // (Ex. an Enter split, which touches two lines) unwind cleanly.
    for i := len(group.actions) - 1; i >= 0; i-- {
        a := group.actions[i]
        switch a.kind {
        case undoEdit:
            if a.line_num >= 0 && a.line_num < len(loadedFile) {
                loadedFile[a.line_num] = a.old_str
            }
        case undoInsert:
            // This action recorded the insertion of a new line at line_num.
            // Undoing it means deleting that line now.
            if a.line_num >= 0 && a.line_num < len(loadedFile) {
                loadedFile = append(loadedFile[:a.line_num], loadedFile[a.line_num+1:]...)
            }
        case undoDelete:
            // This action recorded the removal of a line at line_num.
            // Undoing it means re-inserting old_str back at that position.
            idx := a.line_num
            if idx < 0 { idx = 0 }
            if idx > len(loadedFile) { idx = len(loadedFile) }
            loadedFile = append(loadedFile[:idx], append([]string{a.old_str}, loadedFile[idx:]...)...)
        }
    }

    if len(loadedFile) == 0 { loadedFile = []string{""} }

    cursor_pos = group.cursor_pos
    if cursor_pos[1] < 1 { cursor_pos[1] = 1 }
    if cursor_pos[1] > len(loadedFile) { cursor_pos[1] = len(loadedFile) }
    if cursor_pos[0] < 1 { cursor_pos[0] = 1 }
    sticky_column = cursor_pos[0]

    redo_stack = append(redo_stack, group)
}

// Redo reapplies the most recent undone action group.
func redo() {
    if len(redo_stack) == 0 { return }

    group := redo_stack[len(redo_stack)-1]
    redo_stack = redo_stack[:len(redo_stack)-1]

    lastLine := 1
    for _, a := range group.actions {
        switch a.kind {
        case undoEdit:
            if a.line_num >= 0 && a.line_num < len(loadedFile) {
                loadedFile[a.line_num] = a.new_str
            }
            lastLine = a.line_num + 1
        case undoInsert:
            idx := a.line_num
            if idx < 0 { idx = 0 }
            if idx > len(loadedFile) { idx = len(loadedFile) }
            loadedFile = append(loadedFile[:idx], append([]string{a.new_str}, loadedFile[idx:]...)...)
            lastLine = idx + 1
        case undoDelete:
            if a.line_num >= 0 && a.line_num < len(loadedFile) {
                loadedFile = append(loadedFile[:a.line_num], loadedFile[a.line_num+1:]...)
            }
            lastLine = a.line_num + 1
        }
    }

    if len(loadedFile) == 0 { loadedFile = []string{""} }

    cursor_pos[1] = lastLine
    if cursor_pos[1] < 1 { cursor_pos[1] = 1 }
    if cursor_pos[1] > len(loadedFile) { cursor_pos[1] = len(loadedFile) }
    cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1])) + 1
    sticky_column = cursor_pos[0]

    undo_stack = append(undo_stack, group)
}

func saveForUndo(actions []UndoAction, before_cursor [2]int) {
    if len(actions) == 0 { return }

    undo_stack = append(undo_stack, UndoGroup{actions: actions, cursor_pos: before_cursor})
    redo_stack = redo_stack[:0]

    if len(undo_stack) > undo_max_size {
        undo_stack = undo_stack[len(undo_stack)-undo_max_size:]
    }
}

func saveEditForUndo(line_num int, old_str string, new_str string, before_cursor [2]int) {
    if old_str == new_str { return } // No op edits shouldn't pollute history.
    saveForUndo([]UndoAction{{kind: undoEdit, line_num: line_num, old_str: old_str, new_str: new_str}}, before_cursor)
}

func updateScroll() {
        termSize := returnTermSize()
        termWidth := termSize[0] - 8
        termHeight := termSize[1] - 2

        show_pos = [2]int{cursor_pos[0], cursor_pos[1] - scroll}

        if show_pos[1] > termHeight {
            scroll += show_pos[1] - termHeight
            show_pos[1] = termHeight
        } else if show_pos[1] < 1 {
            scroll += show_pos[1] - 1
            if scroll < 0 { scroll = 0 }
            show_pos[1] = 1
        }

        if cursor_pos[0]-scroll_x > termWidth {
            scroll_x = cursor_pos[0] - termWidth
        } else if cursor_pos[0]-scroll_x < 1 {
            scroll_x = cursor_pos[0] - 1
        }
}

// Helper function to find every match in a string of a string.
// Like strings.Index but it returns the index of each match in that string.
func findAll(string1 string, string2 string) []int {
    matches := []int{}
    for point := 0; point < len(string1); {
        temp := strings.Index(string1[point:], string2)
        if temp != -1 {
            matches = append(matches, point+temp)
            point = point + temp + len(string2)
        } else {
            break
        }
    }
    return matches
}

func hscroll(s string, offset int) string {
    runes := []rune(s)
    if offset >= len(runes) { return "" }
    return string(runes[offset:])
}

// Yet another helper function.
func ask() string {
    for true {
        char, key, err := keyboard.GetKey()
        if err != nil { panic(err) }
        if key == 0 { globalAsk += string(char) } else if char == '\x00' {
            if key == 127 {
                if globalAsk != "" { globalAsk = globalAsk[:len(globalAsk)-1] }
            } else if key == 13 { break } else if key == 32 { globalAsk += " " }
        }
    }
    defer func() { globalAsk, globalPrompt = "", "" }()
    return globalAsk
}

// Yet another helper function.
func yesNoQuestion(msg string) bool {
    globalPrompt = msg
    for true {
        char, _, err := keyboard.GetKey()
        if err != nil  { panic(err)                       }
        if char == 'y' { globalPrompt = "" ; return true  }
        if char == 'n' { globalPrompt = "" ; return false }
    }
    return false
}

func keyPressSystem() {
    if err := keyboard.Open(); err != nil {
        panic(err)
    }
    defer keyboard.Close()

    for {
        char, key, err := keyboard.GetKey()
        if err != nil {
            panic(err)
        }
        mu.Lock()

        globalKey = KeyState{Char: char, Key: key}

        if globalKey.Char == '\x00' && globalKey.Key == 1 {
            fmt.Printf("\x1b[0 q\x1b[?7h\x1b[0;0m\x1b[2J\x1b[H") // <--
            keyboard.Close()
            L.Close()
            os.Exit(0)
        }

        if mode == 0 {
            switch globalKey.Char {
                case '\x00':
                    switch globalKey.Key {
                        case 2: // ctrl+b
                            mode = commandMode
                        //case 4: // ctrl+d
                        //    if err := L.DoString(loadedFile[cursor_pos[1]-1]) ; err != nil { globalPrompt = fmt.Sprintf("Eval error : %v",loadedFile[cursor_pos[1]-1]) ; L.SetTop(0) } else { globalPrompt = "" }
                        case 5: // ctrl+e
                            if cursor_pos[0] > len([]rune(loadedFile[cursor_pos[1]-1])) {
                                cursor_pos[0] = 1
                            } else {
                                cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1]))+1
                            }
                            sticky_column = cursor_pos[0]
                        case 8: // ctrl+backspace
                            before_cursor := cursor_pos
                            old_line := loadedFile[cursor_pos[1]-1]
                            loadedFile[cursor_pos[1]-1] = ""
                            cursor_pos[0] = 1
                            saveEditForUndo(cursor_pos[1]-1, old_line, "", before_cursor)
                        case 9: // tab
                            before_cursor := cursor_pos
                            old_line := loadedFile[cursor_pos[1]-1]
                            insertAtLine(cursor_pos[1]-1,cursor_pos[0]-1,"    ")
                            cursor_pos[0] += 4
                            sticky_column = cursor_pos[0]
                            saveEditForUndo(cursor_pos[1]-1, old_line, loadedFile[cursor_pos[1]-1], before_cursor)
                        case 13: // enter
                            before_cursor := cursor_pos
                            orig := []rune(loadedFile[cursor_pos[1]-1])
                            old_line := string(orig)
                            split_at := cursor_pos[0]-1
                            insertLine(cursor_pos[1])
                            loadedFile[cursor_pos[1]] = string(orig[split_at:])
                            loadedFile[cursor_pos[1]-1] = string(orig[:split_at])
                            first_line_idx := cursor_pos[1]-1
                            cursor_pos[1] += 1
                            if auto_indenting {
                            loadedFile[cursor_pos[1]-1] = strings.Repeat(" ",len(loadedFile[cursor_pos[1]-2])-len(strings.TrimLeft(loadedFile[cursor_pos[1]-2]," "))) + loadedFile[cursor_pos[1]-1]
                            cursor_pos[0] = len(strings.Repeat(" ",len(loadedFile[cursor_pos[1]-2])-len(strings.TrimLeft(loadedFile[cursor_pos[1]-2]," "))))+1
                            } else { cursor_pos[0] = 1 }
                            saveForUndo([]UndoAction{
                                {kind: undoEdit, line_num: first_line_idx, old_str: old_line, new_str: loadedFile[first_line_idx]},
                                {kind: undoInsert, line_num: first_line_idx+1, old_str: "", new_str: loadedFile[first_line_idx+1]},
                            }, before_cursor)
                        case 18: // ctrl+r (for redo)
                            redo()
                            if cursor_pos[0] > len(loadedFile[cursor_pos[1]-1])+1 { cursor_pos[0] = len(loadedFile[cursor_pos[1]-1])+1 }
                        case 21: // ctrl+u (for undo)
                            undo()
                            if cursor_pos[0] > len(loadedFile[cursor_pos[1]-1])+1 { cursor_pos[0] = len(loadedFile[cursor_pos[1]-1])+1 }
                        case 22: // ctrl+v
                            mode = visualMode
                        case 23: // ctrl+w
                            mode = wordMode
                        case 28: // ctrl+\
                            before_cursor := cursor_pos
                            old_line := loadedFile[cursor_pos[1]-1]
                            cursor_pos[0] -= len([]rune(loadedFile[cursor_pos[1]-1])) - len([]rune(strings.TrimLeft(loadedFile[cursor_pos[1]-1]," ")))
                            loadedFile[cursor_pos[1]-1] = strings.TrimLeft(loadedFile[cursor_pos[1]-1]," ")
                            if cursor_pos[0] < 1 { cursor_pos[0] = 1 }
                            saveEditForUndo(cursor_pos[1]-1, old_line, loadedFile[cursor_pos[1]-1], before_cursor)
                        case 32: // space
                            before_cursor := cursor_pos
                            old_line := loadedFile[cursor_pos[1]-1]
                            insertAtLine(cursor_pos[1]-1,cursor_pos[0]-1," ")
                            cursor_pos[0] += 1
                            saveEditForUndo(cursor_pos[1]-1, old_line, loadedFile[cursor_pos[1]-1], before_cursor)
                        case 127: // backspace
                            if cursor_pos[0] > 1 {
                                before_cursor := cursor_pos
                                old_line := loadedFile[cursor_pos[1]-1]
                                line := []rune(loadedFile[cursor_pos[1]-1])
                                line = append(line[:cursor_pos[0]-2], line[cursor_pos[0]-1:]...)
                                loadedFile[cursor_pos[1]-1] = string(line)
                                cursor_pos[0] -= 1
                                sticky_column = cursor_pos[0]
                                saveEditForUndo(cursor_pos[1]-1, old_line, loadedFile[cursor_pos[1]-1], before_cursor)
                            } else if cursor_pos[1] > 1 && cursor_pos[0] == 1 {
                                before_cursor := cursor_pos
                                prev_old := loadedFile[cursor_pos[1]-2]
                                cur_old := loadedFile[cursor_pos[1]-1]
                                cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-2]))+1
                                loadedFile[cursor_pos[1]-2] += loadedFile[cursor_pos[1]-1]
                                loadedFile = append(loadedFile[:cursor_pos[1]-1], loadedFile[cursor_pos[1]:]...)
                                cursor_pos[1] -= 1
                                sticky_column = cursor_pos[0]
                                saveForUndo([]UndoAction{
                                    {kind: undoEdit, line_num: cursor_pos[1]-1, old_str: prev_old, new_str: loadedFile[cursor_pos[1]-1]},
                                    {kind: undoDelete, line_num: cursor_pos[1], old_str: cur_old, new_str: ""},
                                }, before_cursor)
                            }
                        case 65516: // down arrow
                            if cursor_pos[1] < len(loadedFile) {
                                cursor_pos[1] += 1
                                if sticky_column > len([]rune(loadedFile[cursor_pos[1]-1])) {
                                    cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1]))+1
                                } else {
                                    cursor_pos[0] = sticky_column
                                }
                            }
                        case 65517: // up arrow
                            if cursor_pos[1] > 1 {
                                cursor_pos[1] -= 1
                                if sticky_column > len([]rune(loadedFile[cursor_pos[1]-1])) {
                                    cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1]))+1
                                } else {
                                    cursor_pos[0] = sticky_column
                                }
                            }
                        case 65515: // left arrow
                            if cursor_pos[0] > 1 { cursor_pos[0] -= 1 ; sticky_column = cursor_pos[0] }
                        case 65514: // right arrow
                            if cursor_pos[0] <= len([]rune(loadedFile[cursor_pos[1]-1])) { cursor_pos[0] += 1 ; sticky_column = cursor_pos[0] }
                    }
                default:
                    if key == 0 {
                    before_cursor := cursor_pos
                    old_line := loadedFile[cursor_pos[1]-1]
                    insertAtLine(cursor_pos[1]-1,cursor_pos[0]-1,string(globalKey.Char))
                    cursor_pos[0] += 1
                    sticky_column = cursor_pos[0]
                    saveEditForUndo(cursor_pos[1]-1, old_line, loadedFile[cursor_pos[1]-1], before_cursor)
                    }
            }
            show_pos = [2]int{cursor_pos[0],cursor_pos[1]-scroll}
            if show_pos[1] > returnTermSize()[1]-2 {
                scroll += 1 ; show_pos[1] -= 1
            } else if show_pos[1] < 1 {
                scroll -= 1 ; show_pos[1] += 1
            }
            termWidth := returnTermSize()[0] - 8
            if cursor_pos[0]-scroll_x > termWidth {
                scroll_x = cursor_pos[0] - termWidth
            } else if cursor_pos[0]-scroll_x < 1 {
                scroll_x = cursor_pos[0] - 1
            }
        } else if mode == commandMode {
            switch globalKey.Char {
                case 'R':
                    mu.Unlock()
                    globalPrompt = "Insert the string you want to replace>"
                    str1 := ask()
                    globalPrompt = "Insert the string you want to replace it with>"
                    str2 := ask()
                    mu.Lock()
                    for i := 0 ; i < len(loadedFile) ; i++ {
                        loadedFile[i] = strings.ReplaceAll(loadedFile[i],str1,str2)
                    }
                    if cursor_pos[0] > len(loadedFile[cursor_pos[1]-1])+1 { cursor_pos[0] = len(loadedFile[cursor_pos[1]-1])+1 }
                case ':':
                    mu.Unlock()
                    globalPrompt = ":"
                    cmd := ask()
                    mu.Lock()
                    if cmd != "" {
                        if cmd == "guide" || cmd == "man" || cmd == "manual" {
                            //last := loadedFile
                            mu.Unlock()
                            if yesNoQuestion("Are you sure you want to open the manual? This action will erase any non saved changes. (y/n)") {
                            mu.Lock()
                            cursor_pos, show_pos, scroll = [2]int{1,1}, [2]int{1,1}, 0
                            nameFile = "manual.txt"
                            loadedFile = strings.Split(`
        THE MANUAL OF EDITATOR
        ----------------------

(use arrows to move)

Modes:
- Normal mode  | {NRM}  -  The mode were you type in, if you come from vim/neovim think of this like insert mode.
- Command mode | {CMD}  -  The mode that has all the general shortcuts like 'w' for write and 'o' for open.
- Word mode    | {WRD}  -  The mode were you perform actions on the current word (the word under the cursor).
- Visual mode  | {VSL}  -  A mode were you perform actions on the current line and also do copy and paste actions.

(THESE SHORTCUTS ONLY WORK IN NORMAL MODE, TO EXIT A MODE PRESS ENTER)
Command mode = Ctrl+b
Word mode    = Ctrl+w
Visual mode  = Ctrl+v


Normal mode shortcuts:
- Ctrl+e                -  Puts the cursor on the end of the line, if the cursor is already at the end it will put it at the start instead.
- Ctrl+Backspace        -  Deletes the contents of the current line.
- Ctrl+\                -  Remove any whitespaces before the text begins.
- Ctrl+u                -  Undo.
- Ctrl+r                -  Redo.


Command mode shortcuts:
- w                     -  Write to the file.
- e                     -  Go to the end of the file.
- r                     -  Rename the current file.
- R                     -  Replace a string with another one in the whole file.
- f                     -  Find a string in the file (use the up and down arrows to go to the next finding and enter to stop).
- h                     -  Toggle on/off syntax  highlighting.
- -                     -  Toggle 'toggle_mode_after' on/off, when on it automatically goes back to normal mode after doing a command in another mode.
- c                     -  Clear the entire file's contents.
- i                     -  Toggle on/off auto indenting.
- I                     -  Reinitialize the lua plugin system (not reccomended since it's pretty broken).
- :                     -  Open command prompt and output will replace the file content.
- Backspace             -  Delete a range of lines (Ex. 2...7 will delete from line 2 to line 7).


Visual mode:
- up and down arrows    -  Move the line up and down.
- c                     -  Copy from a line to a line.
- p                     -  Paste.
- C                     -  Copy current line.
- Arrows                -  Move the line up/down or indent/dedent.


Word mode:
- c                     -  Toggle upper/lower case on the current word.
- Backspace             -  Remove the current word


I hope you found this guide helpfull!
`,"\n")
                            } else { mu.Lock() }
                        } else {
                            var err error
                            var out []byte
                            if len(strings.Split(cmd, " ")) == 1 {
                                out, err = exec.Command(cmd).Output()
                            } else {
                                out, err = exec.Command(strings.Split(cmd," ")[0],strings.Split(cmd," ")[1:]...).Output()
                            }
                            if err != nil { globalPrompt = "Error." ; break }
                            loadedFile = strings.Split(string(out),"\n")
                            cursor_pos[0], cursor_pos[1], show_pos[0], show_pos[1], scroll = 1, 1, 1, 1, 0
                        }
                    }
                case '-':
                    if toggle_mode_after { toggle_mode_after = false } else { toggle_mode_after = true }
                case 'f': // Find in the file.
                    mu.Unlock()
                    globalPrompt = "string to match>"
                    globalAsk = ask()
                    matches := [][2]int{}

                    for i:=0 ; i<len(loadedFile) ; i++ {
                        l := findAll(loadedFile[i],globalAsk)
                        for k := 0 ; k < len(l) ; k++ { matches = append(matches,[2]int{l[k]+1,i+1}) }
                    }

                    if len(matches) == 0 { globalAsk , globalPrompt = "", "" ; mu.Lock() ; break }

                    matches_count := 0

                    for true {
                        mu.Lock()
                        cursor_pos = matches[matches_count]
                        scroll = cursor_pos[1] - (returnTermSize()[1] - 2)
                        if scroll < 0 { scroll = 0 }
                        show_pos = [2]int{cursor_pos[0], cursor_pos[1] - scroll}
                        mu.Unlock()
                        _, key, err := keyboard.GetKey()
                        if err != nil { panic(err) }
                        if key == 13 {
                            break
                        } else if key == 65517 { // up
                            if matches_count > 0 { matches_count -= 1 }
                        } else if key == 65516 { // down
                            if matches_count < len(matches)-1 { matches_count += 1 }
                        }
                    }

                    globalPrompt = ""
                    globalAsk = ""
                    mu.Lock()
                case 'i':
                    if auto_indenting { auto_indenting = false } else { auto_indenting = true }
                case 'w':
                    mu.Unlock()
                    globalPrompt = fmt.Sprintf("Do you want to write this text to '%v'? (y/anything else)",nameFile)
                    char, _, err := keyboard.GetKey()
                    if err != nil { panic(err) }
                    if char == 'y' {
                        os.WriteFile(nameFile,[]byte(strings.Join(loadedFile,"\n")),0644)
                    }
                    globalPrompt = ""
                    mu.Lock()
                case 'r':
                    mu.Unlock()
                    globalPrompt = "name>"
                    /*for true {
                        char, key, err := keyboard.GetKey()
                        if err != nil { panic(err) }
                        if char == '\x00' && key == 13 {
                            break
                        } else if char == '\x00' && key == 127 {
                            if len(globalAsk) > 0 {
                                globalAsk = globalAsk[:len(globalAsk)-1]
                            }
                        } else if char != '\x00' {
                            globalAsk += string(char)
                        }
                    }*/
                    globalAsk = ask()
                    mu.Lock()
                    if len(globalAsk) != 0 {
                        nameFile = globalAsk
                        tmptype := strings.Split(nameFile,".")
                        fileType = tmptype[len(tmptype)-1]
                    }
                    globalAsk = ""
                    globalPrompt = ""
                case 'o':
                    mu.Unlock()
                    globalPrompt = "file>"
                    /*for true {
                        char, key, err := keyboard.GetKey()
                        if err != nil { panic(err) }
                        if char == '\x00' && key == 13 {
                            break
                        } else if char == '\x00' && key == 127 {
                            if len(globalAsk) > 0 {
                                globalAsk = globalAsk[:len(globalAsk)-1]
                            }
                        } else if char != '\x00' {
                            globalAsk += string(char)
                        }
                    }*/
                    globalAsk = ask()
                    mu.Lock()
                    if len(globalAsk) != 0 {
                        data, err := os.ReadFile(globalAsk)
                        if err != nil {
                            loadedFile = []string{""}
                        } else {
                            loadedFile = strings.Split(string(data),"\n")
                        }
                        nameFile = globalAsk
                        tmptype := strings.Split(nameFile,".")
                        fileType = tmptype[len(tmptype)-1]
                    }
                    globalAsk = ""
                    globalPrompt = ""
                    cursor_pos , show_pos , scroll , scroll_x = [2]int{1,1} , [2]int{1,1} , 0 , 0
                case 'c':
                    mu.Unlock()
                    globalPrompt = "Are you sure you want to clear the contents of this file? (y/n)"
                    for true {
                        char, _, err := keyboard.GetKey()
                        if err != nil { panic(err) }
                        if char == 'y' {
                            mu.Lock()
                            loadedFile = []string{""}
                            cursor_pos , show_pos , scroll , scroll_x = [2]int{1,1} , [2]int{1,1} , 0 , 0
                            break
                        } else if char == 'n' {
                            mu.Lock()
                            break
                        }
                    }
                    globalPrompt = ""
                case 'e':
                    cursor_pos[1] = len(loadedFile)
                    cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1])) + 1
                    termHeight := returnTermSize()[1] - 2
                    scroll = cursor_pos[1] - termHeight
                    if scroll < 0 { scroll = 0 }
                    scroll_x = 0
                    show_pos = [2]int{cursor_pos[0], cursor_pos[1] - scroll}
                case 'g': // Goto line.
                    mu.Unlock()
                    globalPrompt = "Line>"
                    for true {
                        char, key, err := keyboard.GetKey()
                        if err != nil { panic(err) }
                        if char == '\x00' && key == 13 { break } else if char == '\x00' && key == 127 { globalAsk = globalAsk[:len(globalAsk)-1] } else { globalAsk += string(char) }
                    }
                    mu.Lock()
                    n, err := strconv.Atoi(globalAsk)
                    if err == nil && 0 < n && n <= len(loadedFile) {
                        cursor_pos[1] = n
                        cursor_pos[0] = len([]rune(loadedFile[cursor_pos[1]-1])) + 1
                        termHeight := returnTermSize()[1] - 2
                        scroll = cursor_pos[1] - termHeight
                        if scroll < 0 { scroll = 0 }
                        scroll_x = 0
                        show_pos = [2]int{cursor_pos[0], cursor_pos[1] - scroll}
                    }
                    globalAsk = ""
                    globalPrompt = ""
                case 'd': // 
                    before_cursor := cursor_pos
                    index := cursor_pos[1]-1
                    loadedFile = append(loadedFile[:index], append([]string{loadedFile[index]}, loadedFile[index:]...)...)
                    cursor_pos[1]+=1
                    show_pos[1]+=1
                    saveForUndo([]UndoAction{
                        {kind: undoInsert, line_num: index, old_str: "", new_str: loadedFile[index]},
                    }, before_cursor)
                case 'I': // Reinitialize plugins (kinda broken, don't use).
                    initPlugins()
                case 'h': // On/Off highlighting.
                    if highlighting { highlighting = false } else { highlighting = true }
                case 't': // Idk.
                    char,_,err := keyboard.GetKey()
                    if err != nil { panic(err) }
                    cursor_pos[0] = strings.Index(loadedFile[cursor_pos[1]-1],string(char))+1
                    sticky_column = cursor_pos[0]
                case '\x00':
                    switch globalKey.Key { // In these parts im using switch to make it easier in the future to add more keybinds.
                        case 13 :
                            mode = 0
                        case 127:
                            globalPrompt = "lines to delete? (fmt: startLine...endLine) leave blank to cancel>"
                            mu.Unlock()
                            x := ask()
                            mu.Lock()
                            if x != "" {
                                n1, err1 := strconv.Atoi(strings.Split(x,"...")[0])
                                n2, err2 := strconv.Atoi(strings.Split(x,"...")[1])
                                if err1 != nil || err2 != nil { globalPrompt = "Error converting into integer the arguements." ; break }
                                if n2 > len(loadedFile) { n2 = len(loadedFile) }
                                if n1 < 1 { n1 = 1 }
                                loadedFile = append(loadedFile[:n1-1], loadedFile[n2:]...)
                                if cursor_pos[1] > n2 { cursor_pos[1] -= n2-n1+1 ; cursor_pos[0] = 1 } else if cursor_pos[1] >= n1 { cursor_pos[1] = n1 ; cursor_pos[0] = 1 }
                                show_pos = cursor_pos
                            }
                    }
            }
            if toggle_mode_after { mode = 0 }
        } else if mode == wordMode {
            // We get the word start and end indexes.
            line := loadedFile[cursor_pos[1]-1]
            lineLen := len(line)

            wordStartIndex := cursor_pos[0] - 1
            if wordStartIndex > lineLen-1 { wordStartIndex = lineLen - 1 }
            if wordStartIndex < 0 { wordStartIndex = 0 }

            wordEndIndex := cursor_pos[0] - 1
            if wordEndIndex > lineLen-1 { wordEndIndex = lineLen - 1 }
            if wordEndIndex < 0 { wordEndIndex = 0 }

            delimiters := []string{" ",".","[","]","{","}","-","#","@","/","\\","^","?","!","(",")","$","%","*","+","£",":",";",",","<",">","|","&","="}

            if lineLen > 0 {
                for {
                    if contains(delimiters, string(line[wordStartIndex])) {
                        break
                    }
                    wordStartIndex -= 1
                    if wordStartIndex < 0 { wordStartIndex = 0 ; break }
                }

                for {
                    if contains(delimiters, string(line[wordEndIndex])) {
                        break
                    }
                    wordEndIndex += 1
                    if wordEndIndex > lineLen-1 { wordEndIndex = lineLen ; break }
                }
            }

            if globalKey.Char == '\x00' && globalKey.Key == 13 {
                mode = 0
            } else if globalKey.Char == '\x00' && globalKey.Key == 127 {
                loadedFile[cursor_pos[1]-1] = line[:wordStartIndex] + line[wordEndIndex:]
                cursor_pos[0] = wordStartIndex + 1
                if cursor_pos[0] < 1 { cursor_pos[0] = 1 }
                sticky_column = cursor_pos[0]
                show_pos[0] = cursor_pos[0]
            } else if globalKey.Char == 'c' {
                if line[wordStartIndex:wordEndIndex] == strings.ToLower(line[wordStartIndex:wordEndIndex]) {
                    loadedFile[cursor_pos[1]-1] = line[:wordStartIndex] + strings.ToUpper(line[wordStartIndex:wordEndIndex]) + line[wordEndIndex:]
                } else { loadedFile[cursor_pos[1]-1] = line[:wordStartIndex] + strings.ToLower(line[wordStartIndex:wordEndIndex]) + line[wordEndIndex:] }
            }
            if toggle_mode_after { mode = 0 }
        } else if mode == visualMode {
            if globalKey.Char == '\x00' {
                if globalKey.Key == 13 {
                    mode = 0
                } else if globalKey.Key == 65517 { // UP
                    if cursor_pos[1] > 1 {
                        loadedFile[cursor_pos[1]-1], loadedFile[cursor_pos[1]-2] = loadedFile[cursor_pos[1]-2], loadedFile[cursor_pos[1]-1]
                        cursor_pos[1] -= 1 ; updateScroll()
                    }
                } else if globalKey.Key == 65516 { // DOWN
                    if cursor_pos[1] < len(loadedFile) {
                        loadedFile[cursor_pos[1]-1], loadedFile[cursor_pos[1]] = loadedFile[cursor_pos[1]], loadedFile[cursor_pos[1]-1]
                        cursor_pos[1] += 1 ; updateScroll()
                    }
                } else if globalKey.Key == 65514 { // RIGHT
                    loadedFile[cursor_pos[1]-1] = "    " + loadedFile[cursor_pos[1]-1]
                    cursor_pos[0] += 4
                    show_pos[0] += 4
                } else if globalKey.Key == 65515 { // LEFT
                    if loadedFile[cursor_pos[1]-1][:4] == "    " {
                        loadedFile[cursor_pos[1]-1] = loadedFile[cursor_pos[1]-1][4:]
                        if cursor_pos[0] > 4 {
                            cursor_pos[0] -= 4
                            show_pos[0] -= 4
                        }
                    }
                }
            }
        }
        mu.Unlock()
    }
}

func returnTermSize() [2]int {
    width, height, err := term.GetSize(0)
    if err != nil { panic(err) }
    return [2]int{width, height} // Im not returning like width, height but instead in a array because yes
}


func showFile() {
    for {
        cursorVisible(false)
        width, height, err := term.GetSize(0)
        if err != nil {
            panic(err)
        }

        for i := 0; i < height ; i++ {
            printAt(0,i+1,fmt.Sprintf("%v%v",bg_color,strings.Repeat(" ",width)))
        }
        end := height-2+scroll
        if end > len(loadedFile) { end = len(loadedFile) }
        var k string
        visibleWidth := width - 7 // matches the layout offset used by printAt below (k starts at column 7)
        mu.Lock()
        for i := 0 ; i<len(loadedFile[scroll:end]) ; i++ {
            if i == show_pos[1]-1 {
                if highlighting {
                if len(strings.TrimLeft(loadedFile[cursor_pos[1]-1]," ")) == 0 {
                    k = strings.ReplaceAll(hscrollAnsi(colorup(loadedFile[i+scroll],fileType), scroll_x, visibleWidth), " ", "\x1b[38;5;8m•")
                } else {
                    k = hscrollAnsi(colorup(loadedFile[i+scroll],fileType), scroll_x, visibleWidth)
                }
                } else { k = hscroll(loadedFile[i+scroll], scroll_x) }
                //k = colorup(k, fileType) AREYOUSURE
                printAt(0,i+1,fmt.Sprintf("%v%v%v%v \x1b[0m%v %v",cursor_line_count_bg,cursor_line_count_fg,strings.Repeat(" ",5-len(strconv.Itoa(i+1+scroll))),i+1+scroll,bg_color,k))
            } else {
                if highlighting {
                    k = hscrollAnsi(colorup(loadedFile[i+scroll],fileType), scroll_x, visibleWidth)
                } else { k = hscroll(loadedFile[i+scroll],scroll_x) }
                printAt(0,i+1,fmt.Sprintf("\x1b[0;0m%v%v%v%v \x1b[0m%v %v",line_count_bg,line_count_fg,strings.Repeat(" ",5-len(strconv.Itoa(i+1+scroll))),i+1+scroll,bg_color,k))
            }
        }
        mu.Unlock()

        printAt(0,height-1,(fmt.Sprintf("%v%v%v%v","\x1b[0;0m",bottom_bar_bg,strings.Repeat(" ",width),bottom_bar_fg))) //FIND 24;24;24
        printAt(2,height-1,"Editator v0.4")
        printAt(17,height-1,fmt.Sprintf("%v:%v",cursor_pos[1],cursor_pos[0]))
        printAt(34,height-1,fmt.Sprintf("char=%q key=%v", globalKey.Char, globalKey.Key))
        printAt(66,height-1,nameFile)
        if mode == normalMode {
            printAt(56,height-1,"\x1b[1;34m{NRM}")
        } else if mode == commandMode {
            printAt(56,height-1,"\x1b[1;33m{CMD}")
        } else if mode == visualMode {
            printAt(56,height-1,"\x1b[1;32m{VSL}")
        } else if mode == wordMode {
            printAt(56,height-1,"\x1b[1;35m{WRD}")
        }
        printAt(0,height,(fmt.Sprintf("%v%v","\x1b[48;2;54;54;54m",strings.Repeat(" ",width))))
        printAt(0,height,globalPrompt)
        printAt(len(globalPrompt)+1,height,globalAsk)

        printAt(show_pos[0]-scroll_x+7,show_pos[1],"\x1b[0;0m")

        cursorVisible(true)

        flushBuffer()
        time.Sleep(16 * time.Millisecond)
    }
}

func main() {
    printAt(0,0,"\x1b[2J\x1b[?7l") // Clear the screen and stuff.
    fmt.Printf(cursor_type)
    mode = 0
    cursor_pos = [2]int{1,1}
    show_pos   = [2]int{1,1}
    if len(os.Args) > 2 { fmt.Printf("\x1b[1;31mToo many arguements, can only accept one file.\x1b[0;0m\n") ; os.Exit(-1) }
    if len(os.Args) == 2 { nameFile = os.Args[1] } else { nameFile = "test.txt" }
    tmptype := strings.Split(nameFile,".")
    fileType = tmptype[len(tmptype)-1]
    data, err := os.ReadFile(nameFile)
    if err != nil {
        loadedFile = []string{""}
    } else {
        loadedFile = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
    }

    initPlugins()

    go keyPressSystem()
    go showFile()

    // Block main so the program doesn't exit immediately.
    select {}
}