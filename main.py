import terminalUtils as ut
import os
from readchar import readkey, key
import sys



file = ["Welcome to EDITATOR (Alpha)","","Editator is somewhat unstable so don't rely on it too much!","\033[1;31mIf you are new I highly recommend taking a look at this guide","","Shortcuts:",
" - Ctrl+b         = Command mode.",
" - Ctrl+e         = Goes to the end of the line, goes to the start if already at the end.",
" - Ctrl+Backspace = Deletes the content of the current line.",
" - Ctrl+\\        = Removes all of the indentation from the current line.",
" - Ctrl+u         = Undo the last action.",
" - Ctrl+w         = Word mode/operations.",
"",
"Word mode/operations keys:",
" - Backspace = Delete the word under the cursor.",
" - c         = Toggle case of the word under the cursor.",
" - s         = Copy word to the start of the line.",
" - e         = Append word at the end of the line.",
"",
"Command mode keys:",
" - w         = Writes to the file.",
" - r         = Renames the file.",
" - R         = Find and replace a word.",
" - -         = Toggle 'exit command mode after keypress' on/off.",
" - g         = Goto a line.",
" - h         = Toggles on/off the highlight.",
" - c         = Clears the whole file.",
" - u         = Undoes the last action.",
" - o         = Opens a file.",
" - q         = Force quits.",
" - i         = Toggles on/off the auto indentation.",
" - e         = Go to the last line.",
" - backspace = Delete a range of lines (format: startLine&endLine).",
"",
"github: https://github.com/dynamicship31/Editator?tab=readme-ov-file",
"",
"(NOT ALL THE KEYBINDS MIGHT BE PRESENT HERE)",
"",
"Changelog:",
"+ Added horizontal scrolling (FINALLY), although its still buggy",
"+ Added new themes in plugins.py",
"+ Added more customization with colors via plugins",
"+ Added some shortcuts like Ctrl+b & e to go to the end of the file",
"+ Added more shortcuts to the shortcuts guide in the welcome preset file"
]
undo_max_size = 100
file_backups = [[""] for i in range(undo_max_size)]
file_name = "untitled.txt"

if 1 < len(sys.argv) < 3: # Arguement handling.
    try:
        with open(sys.argv[1]) as f:
            file = f.read().split("\n")
        file_name = sys.argv[1]
    except:
        print("Error: Could not open the file provided.")
        exit(0)
elif len(sys.argv) != 1:
    print("Usage: python3 main.py <file path>")

# General variables

undo_count = 0

cursor_pos = [1,1]
scroll = 0
horizontal_scroll = 0

# (colors)
color = "\033[48;2;31;31;31m"
cursor_line_count_bg_color = "\x1b[48;2;65;65;65m"
line_count_bg_color = "\x1b[48;2;34;34;34m"
cursor_line_count_fg = "\x1b[1;33m"
line_count_fg = ""
top_bar_color = ""
bottom_bar_color = ""

# (highlight colors)
default_color        = ""
keyword_color        = ""
special_color        = ""
super_special_color  = ""
string_color         = ""
comment_color        = ""

highlight = True # Turn to false to not have highlighting on by default

toggle_command_after = True # Turn to false if you are a masochist

auto_indent = True

# IO commands
def outputDialogue(msg:str):
    ut.printAt(term_size.lines,0,"\033[38;2;54;54;54m"+" "*term_size.columns)
    ut.printAt(term_size.lines,0,msg)

def inputDialogue(prompt:str) -> str:
    final_output = ""
    ut.printAt(term_size.lines,0,prompt)
    while True:
        thekey = readkey()
        if thekey in ("\n","\r") : break
        elif thekey == key.BACKSPACE:
            final_output+=thekey
        elif thekey.isprintable() and len(thekey) == 1:
            final_output+=thekey
        ut.printAt(term_size.lines,len(prompt),final_output+"\033[38;2;54;54;54m"+" "*(term_size.columns-len(prompt)-len(final_output)))
    return final_output

# File showing stuff

def saveBackup():
    global undo_count
    file_backups[undo_count % undo_max_size] = file[:]
    undo_count += 1

def update_scroll():
    global scroll, horizontal_scroll
    old_scroll = scroll
    if cursor_pos[0] > scroll + term_size.lines - 2:
        scroll = cursor_pos[0] - (term_size.lines - 2)
    if cursor_pos[0] <= scroll:
        scroll = cursor_pos[0] - 1
    if scroll != old_scroll:
        print("\033[2J", end="")

    visible_cols = term_size.columns-8
    if cursor_pos[1] > horizontal_scroll + visible_cols : horizontal_scroll = cursor_pos[1]-visible_cols
    if cursor_pos[1] <= horizontal_scroll : horizontal_scroll = cursor_pos[1]-1

import re

def ansi_len(s: str) -> int: # Return length without counting ansi
    return len(re.sub(r"\033\[[0-9;]*m|\x1b\[[0-9;]*[a-zA-Z]", "", s))

def ansi_strip(s: str) -> str: # Return line without ansi
    return re.sub(r"\033\[[0-9;]*[a-zA-Z]|\x1b\[[0-9;]*[a-zA-Z]", "", s)

def showfile():
    counter = 1 + scroll
    screen_row = 1
    space = " "

    for raw in file[scroll:scroll + term_size.lines - 2]:
        sliced = raw[horizontal_scroll:]

        if highlight:
            import highlight as hl
            i = hl.colorUp(sliced, color, file_name.split(".")[-1], [default_color,keyword_color,special_color,super_special_color,string_color,comment_color])
        else:
            i = sliced

        tmp1 = f"{cursor_line_count_bg_color}{space*(5-len(str(counter)))}{cursor_line_count_fg}{counter} \033[0;0m{color} " + (i.replace(" ", f"\033[38;5;8m•\033[0;0m{color}") if raw[horizontal_scroll:].strip() == "" else i)
        tmp2 = f"{line_count_bg_color}{space*(5-len(str(counter)))}{line_count_fg}{counter} \033[0;0m{color} " + i

        pad = term_size.columns - ansi_len(tmp1 if counter == cursor_pos[0] else tmp2)
        pad = max(0, pad)

        if counter == cursor_pos[0]:
            ut.printAt(screen_row, 0, tmp1 + color + " " * pad + "\033[0;0m")
        else:
            ut.printAt(screen_row, 0, tmp2 + color + " " * pad + "\033[0;0m")

        counter += 1
        screen_row += 1

    for row in range(screen_row, term_size.lines - 1):
        ut.printAt(row, 0, color + " " * term_size.columns + "\033[0;0m")

def showfileOld():
    counter = 1 + scroll
    screen_row = 1
    space = " "

    for i in file[scroll:scroll + term_size.lines - 2]:
        sliced = i[horizontal_scroll:]

        if highlight:
            import highlight as hl
            i = hl.colorUp(i,color,file_name.split(".")[-1],[default_color,keyword_color,special_color,super_special_color,string_color,comment_color])
        else : i = sliced

        tmp1 = f"{cursor_line_count_bg_color}{space*(5-len(str(counter)))}{cursor_line_count_fg}{counter} \033[0;0m{color} " + (i.replace(" ", f"\033[38;5;8m•\033[0;0m{color}") if i.strip()=="" else i)[horizontal_scroll:]
        tmp2 = f"{line_count_bg_color}{space*(5-len(str(counter)))}{line_count_fg}{counter} \033[0;0m{color} " + i[horizontal_scroll:]

        pad = term_size.columns - ansi_len(tmp1 if counter == cursor_pos[0] else tmp2)
        pad = max(0, pad)  # Never negative

        if counter == cursor_pos[0]:
            ut.printAt(screen_row, 0, tmp1 + color + " " * pad + "\033[0;0m")
        else:
            ut.printAt(screen_row, 0, tmp2 + color + " " * pad + "\033[0;0m")

        counter += 1
        screen_row += 1

    for row in range(screen_row, term_size.lines - 1):
        ut.printAt(row, 0, color + " " * term_size.columns + "\033[0;0m")

print("\033[2J",end="")

ut.cursorVisible(False)

show_pos = [1,1]

sticky_column = 1

command = False

def clearCursor():
    ut.printAt(cursor_pos[0]+1,cursor_pos[1]+7," ")
k = " "

plugins_shortcuts = {}

import plugins as plg
plg.init(globals())


while True: # Main loop
    term_size = os.get_terminal_size()

    update_scroll()

    #if cursor_pos[1] > term_size.columns-7 : input("WORKS!")

    show_pos=[cursor_pos[0] - scroll, cursor_pos[1] - horizontal_scroll]

    showfile()

    # Bottom bars
    ut.printAt(term_size.lines-1,0,f"\x1b[48;2;24;24;24m\x1b[38;2;255;255;255m{top_bar_color}  Editator"+" "*(term_size.columns-len("  Editator"))+"\033[0;0m")
    ut.printAt(term_size.lines,0,f"\x1b[48;2;54;54;54m{bottom_bar_color}"+" "*term_size.columns+"\033[0;0m")

    if command : ut.printAt(term_size.lines-1,40,"\033[1;32m{CMD}\033[0;0m")


    ut.printAt(term_size.lines-1,15,cursor_pos)
    ut.printAt(term_size.lines-1,27,[k])

    ut.printAt(show_pos[0], show_pos[1] + 7, "")  # Move real cursor there
    ut.cursorVisible(True)

    print("\033[2 q",flush=True,end="")

    k = readkey()

    if k == "\x01":
        ut.cursorVisible(True)
        print("\033[2J\033[H\033[0 q",end="")
        exit()

    elif command == True: # Command mode?
        cmd_cursor_pos = 0
        if k in ("\n","\r"):
            command = False
        elif k == "w": # Write file
            while True:
                ut.printAt(term_size.lines,0,"\x1b[48;2;54;54;54m"+" "*term_size.columns+"\033[0;0m")
                ut.printAt(term_size.lines,0,"")
                match input("Write?>").lower():
                    case "y" | "yes":
                        with open(file_name,"w") as f:
                            f.write("\n".join(file))
                        break
                    case "n":
                        break
                    case _:
                        continue
        elif k == "o": # Open file
            ut.printAt(term_size.lines,0,"\x1b[48;2;54;54;54m"+" "*term_size.columns+"\033[0;0m")
            ut.printAt(term_size.lines,0,"")
            fpath = input("File?>")
            try:
                with open(fpath, "r") as f:
                    file = f.read().split("\n")
                    file_name = fpath
                cursor_pos = [1,1]
            except:
                ut.printAt(term_size.lines,0,"\033[1;31mError: Could not open file\033[0;0m")

        elif k == "r": # Rename file
            ut.printAt(term_size.lines,0,"\x1b[48;2;54;54;54m"+" "*term_size.columns+"\033[0;0m")
            ut.printAt(term_size.lines,0,"")
            file_name = input("Name?>")
        elif k == "q": # Force quit
            if input("Force quit?>").lower() in ("y","yes") : exit()
        elif k == "h": # Highlight on/off
            highlight = False if highlight == True else True
        elif k == "u": # Undo
            if undo_count > 0:
                undo_count -= 1
                file = file_backups[undo_count % undo_max_size][:]
                if cursor_pos[0] > len(file) : cursor_pos[0] = len(file)
                cursor_pos[1] = len(file[cursor_pos[0]-1])+1
        elif k == "g": # Goto line
            ut.printAt(term_size.lines,0,"\x1b[48;2;54;54;54m"+" "*term_size.columns+"\033[0;0m")
            ut.printAt(term_size.lines,0,"")
            while True:
                goto_line = input("Line number?>")
                try:
                    cursor_pos = [int(goto_line) if 0 < int(goto_line) <= len(file) else cursor_pos[0],1]
                    sticky_column = 1
                    break
                except : continue
        elif k == "c": # Clear file
            ut.printAt(term_size.lines,0,"\x1b[48;2;54;54;54m"+" "*term_size.columns+"\033[0;0m")
            ut.printAt(term_size.lines,0,"")
            if input("Do want to clear the file?(y/anything else)>") == "y" : file = [""];cursor_pos=[1,1]
        elif k == key.BACKSPACE: # Delete from line to line
            ut.printAt(term_size.lines,0,"")
            try:
                usrinput = input("Lines to delete (format: startLine&endLine)>").split("&")
                start = int(usrinput[0]) - 1
                end = int(usrinput[1])

                # Clamp to valid range
                start = max(0, min(start, len(file) - 1))
                end = max(start + 1, min(end, len(file)))

                file[start:end] = []

                # Ensure file is never empty
                if not file:
                    file = [""]

                # Clamp cursor to valid position
                cursor_pos[0] = min(cursor_pos[0], len(file))
                cursor_pos[0] = max(1, cursor_pos[0])
                cursor_pos[1] = min(cursor_pos[1], len(file[cursor_pos[0] - 1]) + 1)
                cursor_pos[1] = max(1, cursor_pos[1])

            except:
                pass
        elif k == "i": # Auto indent on/off
            auto_indent = False if auto_indent else True
        elif k == "R": # Replace (planning on reworking this in the future)
            ut.printAt(term_size.lines,0,"")
            x1 = input("Word to replace>")
            ut.printAt(term_size.lines,0," "*term_size.columns)
            ut.printAt(term_size.lines,0,"")
            x2 = input("Word that replaces it>")
            for index,content in enumerate(file) : file[index] = content.replace(x1,x2)
        elif k == "d":
            file.insert(cursor_pos[0],"")
            file[cursor_pos[0]] = file[cursor_pos[0]-1]
            cursor_pos[0]+=1
        elif k == "\t":
            cursor_pos[1] += 4
            file[cursor_pos[0]-1] = "    "+file[cursor_pos[0]-1]
        elif k == "-":
            toggle_command_after = False if toggle_command_after else True
        elif k == ":": # WIP
            match inputDialogue(":"):
                case "opt":
                    outputDialogue("Optimized!!!")
        elif k == "e":
            cursor_pos = [len(file),1]
        if toggle_command_after == True : command = False

    # Arrows
    elif k == key.UP:
        clearCursor()
        if cursor_pos[0] > 1:
            cursor_pos[0] -= 1
            cursor_pos[1] = min(sticky_column, len(file[cursor_pos[0]-1]) + 1)
    elif k == key.RIGHT:
        clearCursor()
        if cursor_pos[1] <= len(file[cursor_pos[0]-1]) : cursor_pos[1]+=1 ; sticky_column = cursor_pos[1]
    elif k == key.LEFT:
        clearCursor()
        if cursor_pos[1] > 1 : cursor_pos[1]-=1 ; sticky_column = cursor_pos[1]
    elif k == key.DOWN:
        if cursor_pos[0] + 1 <= len(file):
            cursor_pos[0] += 1
            cursor_pos[1] = min(sticky_column, len(file[cursor_pos[0]-1]) + 1)

    # Actual keys
    elif k in ("\r","\n"): # Enter
        saveBackup()
        clearCursor()

        line_before = file[cursor_pos[0]-1]

        file.insert(cursor_pos[0],(" "*(len(line_before)-len(line_before.lstrip(" "))) if auto_indent else "")+file[cursor_pos[0]-1][cursor_pos[1]-1:]) # makes the next line
        file[cursor_pos[0]-1]=file[cursor_pos[0]-1][:cursor_pos[1]-1] # adds to the nextline the things
        cursor_pos[1]=(len(line_before)-len(line_before.lstrip(" ")))+1 if auto_indent else 1
        cursor_pos[0]+=1
    elif k == "\x02": # Ctrl+w (command mode)
        command = True
    elif k == "\x1c": # Ctrl+\ (remove indentation)
        cursor_pos[1] -= max(1,len(file[cursor_pos[0]-1])-len(file[cursor_pos[0]-1].lstrip(" ")))
        file[cursor_pos[0]-1] = file[cursor_pos[0]-1].lstrip(" ")
    elif k == "\x05": # Ctrl+e (goto end/start of the line)
        if cursor_pos[1] == len(file[cursor_pos[0]-1])+1 : cursor_pos[1] = 1 ; sticky_column = cursor_pos[1]
        else : cursor_pos[1] = len(file[cursor_pos[0]-1])+1 ; sticky_column = cursor_pos[1]
    elif k == "\x10":
        second_key = readkey()
        if second_key in plugins_shortcuts:
            plugins_shortcuts[second_key]()
    elif k == "\x08": # Ctrl+Delete (clear line)
        file[cursor_pos[0]-1] = ""
        cursor_pos[1] = 1
    elif k == "\x17":
        usrinput = readkey()
        line = file[cursor_pos[0]-1]
        col = cursor_pos[1] - 1  # 0-indexed
        start = col
        while start > 0 and line[start-1] not in (' ', '\t'):
            start -= 1
        # Search right for end of word
        end = col
        while end < len(line) and line[end] not in (' ', '\t'):
            end += 1

        word = line[start:end]

        match usrinput:
            case key.BACKSPACE:
                # Delete the word
                file[cursor_pos[0]-1] = line[:start] + line[end:]
                cursor_pos[1] = start + 1  # move cursor to where word was
                sticky_column = cursor_pos[1]
            case "c":
                # Toggle case of the word
                if word == word.upper():
                    new_word = word.lower()
                elif word == word.lower():
                    new_word = word.upper()
                else:
                    # Mixed case → go to upper
                    new_word = word.upper()
                file[cursor_pos[0]-1] = line[:start] + new_word + line[end:]
            case "s":
                file[cursor_pos[0]-1] = word+file[cursor_pos[0]-1]
                cursor_pos[1] += len(word)
            case "e":
                file[cursor_pos[0]-1] += word
                cursor_pos[1] += len(word)
            case _:
                0
    elif k == "\x15": # Ctrl+u (Undo)
        if undo_count > 0:
            undo_count -= 1
            file = file_backups[undo_count % undo_max_size][:]
            if cursor_pos[0] > len(file) : cursor_pos[0] = len(file)
            cursor_pos[1] = len(file[cursor_pos[0]-1])+1
    elif k == key.BACKSPACE: # Backspace handling
        saveBackup()
        clearCursor()

        if cursor_pos[1] == 1 and cursor_pos[0] > 1:
            cursor_pos[1] = len(file[cursor_pos[0]-2])+1
            file[cursor_pos[0]-2] += file[cursor_pos[0]-1]
            file.pop(cursor_pos[0]-1)
            cursor_pos[0] -= 1
        elif cursor_pos[1] > 1:
            cursor_pos[1] -= 1
            tmp = list(file[cursor_pos[0]-1])
            tmp.pop(cursor_pos[1]-1)
            file[cursor_pos[0]-1] = "".join(tmp)
            ut.printAt(cursor_pos[0],0," "*term_size.columns)
        sticky_column = cursor_pos[1]
    elif k == "\t": # tabs into spaces
        clearCursor()

        tmp = list(file[cursor_pos[0]-1])
        tmp.insert(cursor_pos[1]-1,"    ")
        file[cursor_pos[0]-1] = "".join(tmp)

        cursor_pos[1]+=4
    elif k.isprintable() and len(k) == 1: # Normal keys
        saveBackup()
        clearCursor()

        tmp = list(file[cursor_pos[0]-1])
        tmp.insert(cursor_pos[1]-1,k)
        file[cursor_pos[0]-1] = "".join(tmp)

        cursor_pos[1]+=1
        sticky_column+=1

    if cursor_pos[1] < 1 : cursor_pos[1] = 1
