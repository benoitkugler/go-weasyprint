import sys

path = sys.argv[1]
with open(path) as f:
    s = ""
    for line in f.readlines():
        parts = line.split("_")
        newline = parts[0]
        for w in parts[1:]:
            if len(w) == 0:
                newline += "_"
            elif w[0] in (",", " ",";"):
                newline += "_" + w
            else:
                newline += w[0].upper() + w[1:]
        if newline[0] == "#":
            newline = "//" + newline[1:]
        newline = newline.replace("'", '"')
        s += newline


with open(path, "w") as f:
    f.write(s)

