""" Generated getters and setters for style object """
import re
import subprocess

SOURCE = "style/properties/initial_values.go"
OUT_1 = "style/properties/accessors.go"
OUT_2 = "style/tree/accessors.go"

RE_PROPERTY = re.compile(r'"(\S+)":\s*([^({\s]+)[({,]')

TEMPLATE_1 = """
func (s {type_name}) Get{prop_cap}() {type_} {{
    return s["{prop}"].({type_})
}}
func (s {type_name}) Set{prop_cap}(v {type_}) {{
    s["{prop}"] = v
}}

"""
TEMPLATE_2 = """
func (s *{type_name}) Get{prop_cap}() {type_} {{
    return s.Get("{prop}").({type_})
}}
func (s *{type_name}) Set{prop_cap}(v {type_}) {{
    s.dict["{prop}"] = v
}}
"""

TEMPLATE_ITF = """
    Get{prop_cap}() {type_} 
    Set{prop_cap}(v {type_})
"""

SPECIAL_VALUES = {
    "zeroPixelsValue": "Value",
    "CurrentColor": "Color",
    "SToV": "Value",
    "FToV": "Value",
}


def camel_case(s: str) -> str:
    out = ""
    for part in s.split("_"):
        out += part.capitalize()
    return out


def parse() -> list:
    metas = []
    with open(SOURCE) as f:
        for line in f.readlines():
            match = RE_PROPERTY.search(line)
            if match:
                prop, type_ = match.groups(1)
                if type_ in SPECIAL_VALUES:
                    type_ = SPECIAL_VALUES[type_]
                metas.append((prop, type_))
    metas = sorted(metas, key=lambda p: p[0])
    return metas


if __name__ == '__main__':
    metas = parse()
    code_1 = """package properties 
    
        // Code generated from properties/initial_values.go DO NOT EDIT

        """
    code_2 = """package tree 
    
        // Code generated from properties/initial_values.go DO NOT EDIT

        import pr "github.com/benoitkugler/go-weasyprint/style/properties"
        
        """
    code_ITF = "type StyleAccessor interface {"
    for prop, type_ in metas:
        prop_cap = camel_case(prop)
        code_1 += TEMPLATE_1.format(type_name="Properties",
                                    prop_cap=prop_cap, prop=prop, type_=type_)
        code_2 += TEMPLATE_2.format(type_name="ComputedStyle",
                                    prop_cap=prop_cap, prop=prop, type_="pr."+type_)
        code_2 += TEMPLATE_2.format(type_name="AnonymousStyle",
                                    prop_cap=prop_cap, prop=prop, type_="pr."+type_)
        code_ITF += TEMPLATE_ITF.format(prop_cap=prop_cap,
                                        prop=prop, type_=type_)

    code_ITF += "}"

    with open(OUT_1, "w") as f:
        f.write(code_1 + code_ITF)
    with open(OUT_2, "w") as f:
        f.write(code_2)

    subprocess.run(["goimports", "-w", OUT_1])
    subprocess.run(["goimports", "-w", OUT_2])
