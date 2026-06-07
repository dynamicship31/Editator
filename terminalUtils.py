def printAt(x:int, y:int, string:str):
    print(f"\033[H\033[{x};{y}H{string}",end="")

def cursorVisible(x:bool):
    if x == False : print(f"\x1b[?25l",end="")
    else : print(f"\x1b[?25h",end="")

def cursorAt(x,y):
    print(f"\033[H\033[{x};{y}H\033[4m\033[{x+1};{y}H\033[0m",end="")
