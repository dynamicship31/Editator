package main

// Prob gonna add some more customizability if I want to in the future.

import (
    "fmt"
    "os"
    "time"
    "github.com/yuin/gopher-lua"
)

var L *lua.LState

func holdOn(L *lua.LState) int {
    numtowait := L.ToInt(1)
    time.Sleep(time.Duration(numtowait) * time.Millisecond)
    return 0
}

func getCursorPos(L *lua.LState) int {
    L.Push(lua.LNumber(cursor_pos[0]))
    L.Push(lua.LNumber(cursor_pos[1]))
    return 2
}

func setColor(L *lua.LState) int {
    name := L.ToString(1)
    rgb  := L.CheckTable(2)

    num1 := rgb.RawGetInt(1)
    num2 := rgb.RawGetInt(2)
    num3 := rgb.RawGetInt(3)

    theString := fmt.Sprintf("\x1b[38;2;%v;%v;%vm",num1,num2,num3)

    switch name {
    case "keyword":
        keywordsColor = theString
    case "type":
        typesColor = theString
    case "built_in_function":
        builtInFunctionsColor = theString
    case "reset_color":
        resetColor = fmt.Sprintf("\x1b[48;2;%v;%v;%vm",num1,num2,num3)
        bg_color = resetColor
    case "loop_and_if":
        loopsAndIfsColor = theString
    case "special_keyword":
        specialKeywordsColor = theString
    case "string":
        stringsColor = theString
    case "comment":
        commentsColor = theString
    case "line_count_fg":
        line_count_fg = theString
    case "line_count_bg":
        line_count_bg = fmt.Sprintf("\x1b[48;2;%v;%v;%vm",num1,num2,num3)
    case "cursor_line_count_fg":
        cursor_line_count_fg = theString
    case "cursor_line_count_bg":
        cursor_line_count_bg = fmt.Sprintf("\x1b[48;2;%v;%v;%vm",num1,num2,num3)
    case "default_color":
        defaultColor = theString
    case "bottom_bar_bg":
        bottom_bar_bg = fmt.Sprintf("\x1b[48;2;%v;%v;%vm",num1,num2,num3)
    case "bottom_bar_fg":
        bottom_bar_fg = fmt.Sprintf("\x1b[38;2;%v;%v;%vm",num1,num2,num3)
    default:
        fmt.Printf("No")
    }

    return 0
}

func getCurrentLine(L *lua.LState) int {
    L.Push(lua.LString(loadedFile[cursor_pos[1]-1]))
    return 1
}

func setLine(L *lua.LState) int {
    line_num := L.ToInt(1)
    str      := L.ToString(2)
    if 0 < line_num && line_num <= len(loadedFile) { loadedFile[line_num-1] = str }
    return 0
}

func getAttribute(L *lua.LState) int {
    str := L.ToString(1)

    switch str {
    case "file_name":
        L.Push(lua.LString(nameFile))
    case "last_char":
        L.Push(lua.LString(string(globalKey.Char)))
    case "last_key":
        L.Push(lua.LString(string(globalKey.Key)))
    }

    return 1
}

func setAttribute(L *lua.LState) int {
    str1 := L.ToString(1)
    str2 := L.ToString(2)

    switch str1 {
    case "file_name":
        nameFile = str2
    case "cursor_type":
        if str2 == "block" {
        fmt.Printf("\x1b[2 q")
        } else if str2 == "blinking_block" {
        fmt.Printf("\x1b[1 q")
        } else if str2 == "vertical_line" {
        fmt.Printf("\x1b[6 q")
        } else if str2 == "blinking_vertical_line" {
        fmt.Printf("\x1b[5 q")
        } else if str2 == "horizontal_line" {
        fmt.Printf("\x1b[4 q")
        } else if str2 == "blinking_horizontal_line" {
        fmt.Printf("\x1b[3 q")
        }
    }

    return 0
}

func appendLineAt(L *lua.LState) int {
    index := L.ToInt(1)
    insertLine(index)
    return 0
}

var homedir string

func loopPlugins() {
    for true {
    if err := L.DoFile(homedir+"/.config/editator/loop.elua"); err != nil { fmt.Println("plugin loop error:",err) }
    }
}

func initPlugins() {
    L = lua.NewState()
    //defer L.Close()

    var err error
    homedir , err = os.UserHomeDir()
    if err != nil { globalPrompt = "Error finding the home directory." ; return }

    L.SetGlobal("get_cursor_pos",L.NewFunction(getCursorPos))
    L.SetGlobal("set_color",L.NewFunction(setColor))
    L.SetGlobal("set_line",L.NewFunction(setLine))
    L.SetGlobal("get_current_line",L.NewFunction(getCurrentLine))
    L.SetGlobal("get_attribute",L.NewFunction(getAttribute))
    L.SetGlobal("append_line_at",L.NewFunction(appendLineAt))
    L.SetGlobal("wait",L.NewFunction(holdOn))
    L.SetGlobal("set_attribute",L.NewFunction(setAttribute))

    if _, err := os.Stat(homedir+"/.config/editator/plugins.elua") ; err == nil {
        if err := L.DoFile(homedir+"/.config/editator/plugins.elua"); err != nil { panic(err) }
    }

    if _, err := os.Stat(homedir+"/.config/editator/loop.elua") ; err == nil {
        go loopPlugins()
    }
}
