package main

import ("fmt";"strings")

// colorup version 0.3

// Might migrate to lua (maybe).

func contains(arr []string, target string) bool { // Helper function to see if a element is in an array.
    for _, v := range arr {
        if v == target {
            return true
        }
    }
    return false
}

var resetColor              string = "\x1b[48;2;31;31;31m"
var defaultColor            string = "\x1b[39m"
var keywordsColor           string = "\x1b[38;2;197;134;192m" //"\x1b[1;33m"
var typesColor              string = "\x1b[38;2;18;180;200m"
var builtInFunctionsColor   string = "\x1b[1;34m"
var loopsAndIfsColor        string = "\x1b[1;31m"
var specialKeywordsColor    string = "\x1b[1;35m"
var stringsColor            string = "\x1b[38;2;100;170;100m"
var commentsColor           string = "\x1b[38;5;8m"

func goodSplit(content string, dividers []string, comment string) []string {
    tokens := []string{}
    token := ""
    skip := false

    runes := []rune(content)

    for b := 0; b < len(runes); b++ {
        i := string(runes[b])
        if skip {
            skip = false
        } else if contains(dividers,i) && b+1 < len(runes) && contains(dividers,i) && len(comment) > 1 {
            h := i+string(runes[b+1]) // h is the first and the second divider in a string
            if h == comment {
                if token != "" { tokens = append(tokens,token) }
                token = ""
                tokens = append(tokens, h)
                skip = true
            } else {
                if token != "" { tokens = append(tokens, token) }
                token = ""
                tokens = append(tokens, i)
            }
        } else if contains(dividers, i) {
            if token != "" { tokens = append(tokens, token) }
            token = ""
            tokens = append(tokens, i)
        } else {
            token += i
        }
    }
    if token != "" { tokens = append(tokens, token) }

    return tokens
}

func generalHighlight(content string, keywords []string, types []string, builtInFunctions []string, loopsAndIfs []string, specialKeywords []string, comment string, case_sensitivity bool) string {
    tokens := goodSplit(content, []string{"\"","#","(",")","[","]","{","}",
    "-","+","*","/","%","&","=","£","$","'","!","^","<",">","|",":",";",",","."," "},comment)

    parts := []string{}

    in_string   := false
    escape      := false
    string_char := "__"

    for i := 0 ; i < len(tokens) ; i++ {
        token := tokens[i]
        tokenCmp := token

        if case_sensitivity == false { tokenCmp = strings.ToLower(token) }

        if in_string {
            parts = append(parts,fmt.Sprintf("%v%v%v%v",resetColor,stringsColor,token,defaultColor))
            if escape {
                escape = false
            } else if token == "\\" {
                escape = true
            } else if token == string_char && escape != true {
                in_string = false
            }
        } else if token == comment {
            parts = append(parts,fmt.Sprintf("%v%v%v%v",resetColor,commentsColor,strings.Join(tokens[i:],""),defaultColor))
            break
        } else if token == "\"" || token == "'" {
            in_string = true
            string_char = token
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor,stringsColor,token,defaultColor))
        } else if contains(keywords, tokenCmp) {
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor, keywordsColor, token, defaultColor))
        } else if contains(types, tokenCmp) {
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor, typesColor, token, defaultColor))
        } else if contains(builtInFunctions, tokenCmp) {
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor, builtInFunctionsColor, token, defaultColor))
        } else if contains(loopsAndIfs, tokenCmp) {
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor, loopsAndIfsColor, token, defaultColor))
        } else if contains(specialKeywords, tokenCmp) {
            parts = append(parts, fmt.Sprintf("%v%v%v%v",resetColor, specialKeywordsColor, token, defaultColor))
        } else {
            parts = append(parts, fmt.Sprintf("\x1b[0;0m%v%v%v%v",resetColor, defaultColor, token, defaultColor))
        }
    }

    return defaultColor+strings.Join(parts,"")
}

func colorup(content string, file_type string) string {
    var case_sens    bool = true
    var thingys      [][]string
    var commentchar  string

/*
FMT FOR COLORUP
1. KEYWORDS
2. TYPES
3. BUILT IN FUNCTIONS
4. LOOPS AND IFS STUFF
5. SPECIAL KEYWORDS

    |
    V

[][]string {
[]string{KEYWORDS...},
[]string{TYPES...},
[]string{BUILT IN FUNCTIONS...}
[]string{LOOPS AND IFS...}
[]string{SPECIAL KEYWORDS...},
}
*/

    switch file_type {
        case "py":
            thingys = [][]string{
            []string{"return","def"},
            []string{"int","str","bool","float"},
            []string{"input","enumerate","len"},
            []string{"while","if","else","for","in"},
            []string{"print","and","or","not","is"},
            }
            commentchar = "#"
        case "go":
            thingys = [][]string{
            []string{"func","go","var","const","return"},
            []string{"int","int8","int16","int32","int64","float","string","bool"},
            []string{"if","else","switch","case","select","default","for"},
            []string{"Printf","PrintLn","Print","import"},
            []string{"package","true","false"},
            }
            commentchar = "//"
        case "elua":
            thingys = [][]string {
            []string{"function","end","return","if","then","else","elseif","for","do","in","until","while"},
            []string{"boolean","thread","table","number","string","userdata","break"},
            []string{"goto","true","false","and","not","or"},
            []string{"set_color","get_cursor_pos"},
            []string{"nil","local"},
            }
            commentchar = "--"
        case "lua":
            thingys = [][]string{
            []string{"if","then","else","elseif","end","for","while","do","repeat","until","function","local","return","break"},
            []string{"number","string","boolean","table","function","nil"},
            []string{"print","pairs","ipairs","type","tostring","tonumber","table","math"},
            []string{"if","for","while","repeat"},
            []string{"function","local","require"},
            }
            commentchar = "--"
        case "c":
            thingys = [][]string{
            []string{"int","char","float","double","if","else","switch","case","break","continue","return","for","while","do","struct","typedef","enum","static","extern","sizeof"},
            []string{"int","char","float","double","void","struct","union","enum","pointer"},
            []string{"printf","scanf","malloc","free","memcpy","strlen","fopen","fclose"},
            []string{"if","else","switch","for","while","do-while","goto"},
            []string{"include","define","ifdef","ifndef","pragma"},
            }
            commentchar = "//"
        case "java":
            thingys = [][]string{
            []string{"class","public","private","protected","static","final","if","else","switch","case","break","continue","return","for","while","do","new","this","extends","implements"},
            []string{"int","long","short","byte","float","double","char","boolean","String","Object"},
            []string{"System.out.println","Math","Arrays","String"},
            []string{"if","else","switch","for","while","do-while"},
            []string{"package","import","throws","try","catch","finally","interface","enum"},
            }
            commentchar = "//"
        case "for", "f90", "f99":
            thingys = [][]string{
            []string{"program","end","if","then","else","elseif","do","while","function","subroutine","module","use","implicit","none","select","case"},
            []string{"integer","real","double precision","logical","character"},
            []string{"print","read","sum","max","min","abs","sin","cos"},
            []string{"goto"},
            []string{"program","module","contains"},
            }
            commentchar = "!"
            case_sens = false
        case "test":
            thingys = [][]string{
            []string{"Func","Init","Globals","Let","Class","Return"},
            []string{"Num","String","Bool","List"},
            []string{"Print","Format","Exit"},
            []string{"If","Else","While"},
            []string{"Include","Push","Pop","Peek"},
            }
            commentchar = "//"
        default:
            return defaultColor+content
    }

    return generalHighlight(content, thingys[0], thingys[1], thingys[2], thingys[3], thingys[4], commentchar, case_sens)
}