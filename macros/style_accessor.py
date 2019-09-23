""" Generated getters and setters for style object """
import re

SOURCE = "style/properties/initial_values.go"
OUT = "style/properties/accessors.go"

RE_PROPERTY = re.compile(r'"(\S+)":\s*([^({\s]+)[({,]')

TEMPLATE = """
func (s Properties) Get{prop_cap}() {type_} {{
    return s["{prop}"].({type_})
}}
func (s Properties) Set{prop_cap}(v {type_}) {{
    s["{prop}"] = v
}}
"""

SPECIAL_VALUES = {
    "zeroPixelsValue": "Value",
    "CurrentColor": "Color",
    "SToV": "Value",
    "FToV": "Value",
}


def camel_case(s: str):
    out = ""
    for part in s.split("_"):
        out += part.capitalize()
    return out

def parse():
    metas = []
    with open(SOURCE) as f:
        for line in f.readlines():
            match = RE_PROPERTY.search(line)
            if match:
                prop, type_ = match.groups(1)
                if type_ in SPECIAL_VALUES:
                    type_ = SPECIAL_VALUES[type_]
                metas.append((prop, type_))
    return metas

if __name__ == '__main__':
    metas = parse()
    code = """package properties 
    
        // autogenerated from initial_values.go
        """
    for prop, type_ in metas:
        prop_cap = camel_case(prop)
        code += TEMPLATE.format(prop_cap=prop_cap, prop=prop, type_=type_)

    with open(OUT, "w") as f:
        f.write(code)