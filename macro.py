import re
import sys

refunc = re.compile("^(def )")

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
        newline = newline.replace("True", "true")
        newline = newline.replace("False", "false")
        newline = newline.replace(" and ", " && ")
        newline = newline.replace(" or ", " || ")
        newline = newline.replace("# ", "// ")
        newline = refunc.sub("func ", newline)
        s += newline


with open(path, "w") as f:
    f.write(s)
