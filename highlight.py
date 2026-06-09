# COLORUP HIGHLIGHTING SYSTEM
# Colorup is still not complete and doesn't support multiline comments and strings for now.

"""
SUPPORTED LANGUAGES

Python, Javascript/Typescript, Haskell, Porth,
C, Lua, Go, Fortran, Bash

FUTURE LANGUAGES TO SUPPORT

Php, Jai, Rust, Ada, Pascal, Hare,
QbeBackend IR, LLVM IR
"""

BG = "\033[48;2;31;31;31m"

def reset(bg=BG):
    return f"\033[0;0m{bg}"

def goodsplit(dividers: list, content: str):
    tokens = []
    token = ""
    skip = False

    for b, i in enumerate(content):
        if skip : skip = False
        elif (i in dividers and i == "/" and b+1 < len(content) and content[b+1] == "/") or (i in dividers and i == "-" and b+1 < len(content) and content[b+1] == "-"):
            if token != "" : tokens.append(token)
            token = ""
            tokens.append(i+content[b+1])
            skip = True
        elif i in dividers:
            if token != "" : tokens.append(token)
            token = ""
            tokens.append(i)
        else:
            token += i
    if token != "": tokens.append(token)

    return tokens


def generalColorUp(x:str, bg=BG, keywords=[], special_keywords=[], super_special_keywords=[], comment="//", case_sensitivity=True,colors=["","","","","",""]):
    tokens = goodsplit(list("\"#()[]{}-+*/%&=£$'!^<>|:;,. "), x)
    parts = []

    in_string = False
    escape = False
    string_char = ""

    for idx, token in enumerate(tokens):
        if in_string:
            parts.append(f"\033[38;2;100;170;100m{colors[4]}{token}{reset(bg)}")
            if escape:
                escape = False  # this char was escaped, skip close check
            elif token == "\\":
                escape = True
            elif token == string_char:
                in_string = False

        elif token == comment:
            parts.append(f"\033[38;5;8m{colors[5]}{''.join(tokens[idx:])}{reset(bg)}")
            break
        elif token in ('"', "'"):
            in_string = True
            string_char = token
            parts.append(f"\033[38;2;100;170;100m{colors[4]}{token}{reset(bg)}")
        elif (token if case_sensitivity else token.lower()) in keywords:
            if (token if case_sensitivity else token.lower()) in super_special_keywords : parts.append(f"\033[38;2;18;180;200m{colors[3]}{token}{reset(bg)}")
            elif (token if case_sensitivity else token.lower()) in special_keywords : parts.append(f"\033[1;31m{colors[2]}{token}{reset(bg)}")
            else:
                parts.append(f"\033[1;33m{colors[1]}{token}{reset(bg)}")
        elif token.isdigit():
            parts.append(f"\033[38;5;81m{token}{reset(bg)}")
#        elif not token in "\"#()[]{}-+*/%&=£$'!^<>|:;,. ":
#            parts.append(f"\033[38;2;156;220;254m{token}{reset(bg)}")
        else : parts.append(colors[0]+token)

    return "".join(parts)

def colorUp(x:str, bg=BG,file_type:str = "txt", colors:list=["","","","","",""]):
    case_sens = True
    match file_type:
        case "py":
            tmp_keywords = [
            ["len",'print','enumerate','False', 'None', 'True', 'and', 'as', 'assert', 'async', 'await', 'break', 'case', 'class', 'continue', 'def', 'del', 'elif',
            'else', 'except', 'finally', 'for', 'from', 'global', 'if', 'import', 'in', 'is', 'lambda', 'match', 'nonlocal', 'not', 'or', 'pass', 'raise', 'return',
            'try', 'while', 'with', 'yield','str','int']
            ,["if","for","while","True","False","None","and","in","not","is","elif","else","continue","pass"]
            ,["import","try","except","print","enumerate","len"]]
            commentchar = "#"
        case "c" | "h":
            tmp_keywords = [
            ["include","auto","break","case","char","const","continue","default","do","double","else","enum","extern","float","for","goto","if","inline","int","long",
            "register","restrict","return","short","signed","sizeof","static","struct","switch","typedef","union","unsigned","void","volatile","while","_Alignas",
            "_Alignof","_Atomic","_BitInt","_Bool","_Complex","_Decimal128","_Decimal32","_Decimal64","_Generic","_Imaginary","_Noreturn","_Static_assert","_Thread_local"]
            ,["if","for","while","switch","case","extern"]
            ,["include","int","char","void","goto","struct"]]
            commentchar = "//"
        case "porth":
            tmp_keywords = [
            ["include","proc","in","end","dup","do","while","puts","if","print","swap","drop","over","rot","max","divmod","shr","shl","or","and","not","cast","int","bool","ptr",
            "true","false","const","addr"],
            ["if","do","while","end","or","and","not","true","false"],
            ["include","int","bool","ptr","addr","puts","print"]
            ]
            commentchar = "//"
        case "js" | "ts":
            tmp_keywords = [
            ['await','break','case','catch','class','const','continue','debugger','default','delete',
            'do','else','enum','export','extends','false','finally','for','function','if',
            'import','in','instanceof','new','null','return','super','switch','this','throw',
            'true','try','typeof','var','void','while','with','yield'],
            ["true","try","false","in","if","continue","break","else","do","delete","case","catch","continue","switch","for","throw","yield","while"],
            ["import","export","var","const","class"]
            ]
            commentchar = "//"
        case "hs":
            tmp_keywords = [
            ['case','class','data','default','deriving','do','else','foreign','if','import',
            'in','infix','infixl','infixr','instance','let','module','newtype','of','then',
            'type','where','as','qualified','hiding','family','role','pattern','proc','rec',
            'mdo','Maybe','Integer','Bool','IO','Int','putStr','putStrLn'],
            ["do","else","if","then","in","case","class"],
            ["Maybe","Integer","IO","Bool","Int"]
            ]
            commentchar = "--"
        case "go":
            tmp_keywords = [
            ['break','case','chan','const','continue','default','defer','else','fallthrough','for',
            'func','go','goto','if','import','interface','map','package','range','return',
            'select','struct','switch','type','var','int','int64','int32','int16','int8','rune','bool','float'],
            ["for","if","else","goto","select","switch","case","break","continue","import"],
            ["package","struct","rune","int8","int16","int32","int64","bool","float"]
            ]
            commentchar = "//"
        case "lua":
            tmp_keywords = [
            ['and','break','do','else','elseif','end','false','for','function','goto',
            'if','in','local','nil','not','or','repeat','return','then','true',
            'until','while',"print"],
            ["break","if","in","and","do","else","elseif","end","false","for","goto","then","true","not","or","while","until"],
            ["nil","local"]
            ]
            commentchar = "--"
        case "for" | "f90":
            tmp_keywords = [
            ["abstract", "allocatable", "allocate", "assignment", "associate","asynchronous", "backspace", "bind", "block", "blockdata",
            "call", "case", "class", "close", "codimension","common", "complex", "concurrent", "contains", "continue",
            "critical", "cycle", "data", "deallocate", "default","deferred", "dimension", "do", "doubleprecision", "else",
            "elseif", "elsewhere", "end", "endassociate", "endblock","endcritical", "endenum", "endfile", "endforall", "endfunction",
            "endif", "endinterface", "endmodule", "endprocedure", "endprogram","endselect", "endsubmodule", "endsubroutine", "endteam", "endtype",
            "endwhere", "entry", "enum", "enumerator", "equivalence","errorstop", "event", "exit", "extends", "external",
            "file", "final", "flush", "forall", "format","function", "generic", "go", "goto", "if",
            "implicit", "import", "in", "include", "inout","integer", "intent", "interface", "intrinsic", "lock",
            "logical", "module", "namelist", "none", "non_intrinsic","nopass", "only", "open", "operator", "optional",
            "out", "parameter", "pass", "pause", "pointer","print", "private", "procedure", "program", "protected",
            "public", "pure", "read", "real", "recursive","result", "return", "rewind", "save", "select",
            "sequence", "stop", "submodule", "subroutine", "sync","target", "team", "then", "type", "unlock",
            "use", "value", "volatile", "wait", "where","while", "write"],
            ["open","goto","read","pass","include","exit","call","continue"],
            ["implicit","integer","logical","character","print","write","doubleprecision","file"]
            ]
            commentchar = "!"
            case_sens = False
        case "sh" | "bash":
            tmp_keywords = [
            ["if", "then", "else", "elif", "fi", "case", "esac", "for", "select", "while",
            "until", "do", "done", "in", "function", "time",
            "coproc", "string", "integer", "array", "associative_array", "function", "readonly", "exported", "local",
            "nameref", "alias", "bg", "bind", "break", "builtin", "caller", "cd", "command", "compgen",
            "complete", "compopt", "continue", "declare", "dirs", "disown", "echo", "enable", "eval", "exec",
            "exit", "export", "fc", "fg", "getopts", "hash", "help", "history", "jobs", "kill",
            "let", "local", "logout", "mapfile", "popd", "printf", "pushd", "pwd", "read", "readarray",
            "readonly", "return", "set", "shift", "shopt", "source", ".", "suspend", "test", "[",
            "times", "trap", "type", "typeset", "ulimit", "umask", "unalias", "unset", "wait"],
            ["until","if","fi","then","for","while","do","in","done"],
            ["alias","jobs","kill","umask","unalias","unset","wait","exit","complete"]
            ]
            commentchar = "#"
        case "bf":
            tmp_keywords = [["<>-+.[],"],[],[]]
            commentchar = "comment"
        case _:
            return colors[0]+x

    return generalColorUp(x,bg,tmp_keywords[0],tmp_keywords[1],tmp_keywords[2],commentchar,case_sens,colors)