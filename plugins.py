bg_color = "\033[48;2;18;18;18m"

def batman(extern_vars):
    extern_vars["color"] = bg_color
    extern_vars["cursor_line_count_bg_color"] = bg_color
    extern_vars["line_count_bg_color"] = bg_color
    extern_vars["cursor_line_count_fg"] = "\x1b[38;2;220;220;0m"
    extern_vars["line_count_fg"] = "\x1b[38;2;55;55;55m"

def blueBerry(extern_vars):
    extern_vars["color"] = "\033[48;2;20;20;50m"
    extern_vars["default_color"] = "\033[38;2;167;216;245m"
    extern_vars["line_count_fg"] = "\033[38;2;67;116;145m"
    extern_vars["cursor_line_count_fg"] = "\033[38;2;100;100;150m"
    extern_vars["cursor_line_count_bg_color"] = "\033[48;2;18;18;18m"
    extern_vars["keyword_color"] = "\033[38;2;214;240;255m"

def cherry(extern_vars):
    extern_vars["color"] = "\033[48;2;30;30;30m"
    extern_vars["cursor_line_count_bg_color"] = "\033[48;2;60;8;20m"
    extern_vars["line_count_bg_color"] = "\033[48;2;30;30;30m"
    extern_vars["cursor_line_count_fg"] = "\033[38;2;247;138;168m"
    extern_vars["line_count_bg_color"] = "\033[48;2;101;26;49m"
    extern_vars["line_count_fg"] = "\033[38;2;0;0;0m"
    extern_vars["keyword_color"] = "\033[1;31m"

def init(extern_vars):
    theme = "default"

    match theme:
        case "batman":
            batman(extern_vars)
        case "blueBerry":
            blueBerry(extern_vars)
        case "cherry":
            cherry(extern_vars)
        case "default":
            return 0;
