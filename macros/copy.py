import re
import sys

import style_accessor

""" Generated the Copy methods, recursively """

sys.path.append("macros")


SOURCE = "style/properties/types.go"
OUT = "style/properties/generated.go"

RE_TYPEDEF = re.compile(r"^type (\S+) ((struct {)|\[\d?\](\S+)|[\w.]+)")

RE_ARRAY = re.compile(r"^\[\d+\](\S+)")

RE_FIELDS = re.compile(r"(\w+)\s*([\[\]\w]+)?")


TEMPLATE_HEADER_COPY = """func (v {type_}) Copy() CssProperty {{ {body}"""
TEMPLATE_HEADER_PRIVATE_COPY = """func (v {type_}) copy() {type_} {{ {body}"""

TEMPLATE_BODY_SLICE_TYPE = """
    out := make({type_}, len(v))
    for i, k := range v {{
        out[i] = {subtype_copy}
    }}
    return out
}}
"""

TEMPLATE_BODY_STRUCT_TYPE = """
    out := v
    {fields}
    return out
}}
"""

TEMPLATE_HEADER_IS_NONE = """func (v {type_}) IsNone() bool {{
    {body}
}}
"""

# type_ : kind, subtype (for array and slice) or {field: subtype} for struct
# possible kind: array, slice, struct, ident
TYPES = {
    "string": ("builtin", ""),
    "int": ("builtin", ""),
    "float32": ("builtin", ""),
    "bool": ("builtin", ""),
    "uint8": ("builtin", ""),
    "parser.Color": ("builtin", ""),
    "parser.Token": ("builtin", "slice"),
    "Set": ("builtin", "set"),
    "CssProperty": ("builtin", "interface"),
    "ValidatedProperty": ("builtin", "interface")
}

ZERO_VALUES = {
    "string": '""',
    "int": "0",
    "float32": "0.",
    "bool": "false",
    "uint8": "0"
}


def analyse_struct(lines):
    out = {}
    for line in lines:
        if not line.strip() or line.strip().startswith("//"):
            continue
        match = RE_FIELDS.search(line)
        field = match.group(1)
        if match.group(2):
            type_ = match.group(2)
        else:
            type_ = field
        out[field] = analyse_type(type_)
    return out


def analyse_type(value: str):
    match_array = RE_ARRAY.search(value)
    if match_array:
        next_type = match_array.group(1)
        meta = analyse_type(next_type)
        return ("array", meta)
    elif value.startswith("[]"):
        meta = analyse_type(value[2:])
        return ("slice", meta)
    elif value.startswith("struct"):
        lines = value.strip().split("\n")[1:]
        return ("struct", analyse_struct(lines))
    else:
        return ("ident", value)


def isValueType(details):
    kind, meta = details
    if meta in ("interface", "pointer"):
        return False
    elif kind == "ident":
        return isValueType(TYPES[meta])
    elif kind == "slice":
        return False
    elif kind == "array":
        return isValueType(meta)
    elif kind == "struct":
        for _, field_type in meta.items():
            if not isValueType(field_type):
                return False
        return True
    elif kind == "builtin":
        return meta not in ("slice", "set")


with open(SOURCE) as f:
    code = """package properties 

    // autogenerated from types.go
    """
    in_struct = False
    struct_lines = []
    type_ = ""
    for line in f.readlines():
        if in_struct and line == "}\n":
            TYPES[type_] = analyse_type(content)
            in_struct = False
            content = ""

        if in_struct:
            content += line

        match = RE_TYPEDEF.search(line)
        if match:
            type_ = match.group(1)
            content = match.group(2)
            if content.startswith("struct"):
                in_struct = True
                content += "\n"
            else:
                TYPES[type_] = analyse_type(content)

needed_types = sorted(set(type_ for _, type_ in style_accessor.parse()))

need_private_copy = {"CustomProperty", "RadialGradient", "LinearGradient", "SContentProps"}


def generate_body_copy_method(type_: str) -> str:
    kind, details = TYPES[type_]
    if kind == "builtin" or details == "interface":
        return " }\n"
    if isValueType((kind, details)):
        return " return v }\n"
    elif kind in ("slice", "array"):
        subtype, subdetails = details
        if isValueType(details):
            subtype_copy = "k"
        else:
            if subtype != "ident":
                raise ValueError(
                    "%s : slice subtype need copy() method" % type_)
            need_private_copy.add(subdetails)
            if TYPES[subdetails][1] == "interface":
                subtype_copy = "k.copyAs%s()" % subdetails
            else:
                subtype_copy = "k.copy()"

        return TEMPLATE_BODY_SLICE_TYPE.format(type_=type_, subtype_copy=subtype_copy)
    elif kind == "struct":
        fields = ""
        for field, subtype in details.items():
            if not isValueType(subtype):
                subkind, subtypename = subtype
                if subkind != "ident":
                    raise ValueError(
                        "field %s in struct must be a defined type" % str(subtype))
                if TYPES[subtypename][1] == "interface":
                    subtype_copy = "copyAs%s()" % subtypename
                else:
                    subtype_copy = "copy()"

                fields += f"out.{field} = v.{field}.{subtype_copy}\n"
                need_private_copy.add(subtypename)
        return TEMPLATE_BODY_STRUCT_TYPE.format(type_=type_, fields=fields)


need_is_none = set(TYPES.keys())


def generate_zero_method(type_):
    kind, details = TYPES[type_]
    if kind == "struct":
        if isValueType((kind, details)):
            body = f"return v == {type_}{{}}"
        else:
            tests = []
            for field, subtype in details.items():
                subtypekind, subtypename = subtype

                if subtypename in ZERO_VALUES:
                    tests.append(f" v.{field} == {ZERO_VALUES[subtypename]} ")
                elif isValueType(subtype):
                    tests.append(f" v.{field} == {subtypename}{{}} ")
                elif subtypekind in ("slice", "interface"):
                    tests.append(f" v.{field} == nil ")
                elif subtypekind == "ident":
                    if TYPES[subtypename][0] in ("slice", "interface"):
                        tests.append(f" v.{field} == nil ")
                else:
                    tests.append(f" v.{field}.IsNone() ")
                    need_is_none.add(subtypename)
            body = "return " + " && ".join(tests)
    elif kind == "array":
        body = f" return v ==  {type_}{{}}"
    else:
        return
    return TEMPLATE_HEADER_IS_NONE.format(type_=type_, body=body)


code = """package properties 
    
        // autogenerated from initial_values.go
        
        """

done_zero = set()
for type_ in needed_types:
    if TYPES[type_][1] == "interface":
        continue

    body = generate_body_copy_method(type_)

    code += TEMPLATE_HEADER_COPY.format(type_=type_, body=body) + "\n"
    c = generate_zero_method(type_)
    if c is not None:
        code += c
    done_zero.add(type_)

done = set()
while need_private_copy:
    type_ = need_private_copy.pop()
    if not type_ in done:
        kind, det = TYPES[type_]
        if not (det == "interface" or kind == "builtin"):
            body = generate_body_copy_method(type_)
            code += TEMPLATE_HEADER_PRIVATE_COPY.format(
                type_=type_, body=body) + "\n"

    done.add(type_)

while need_is_none:
    type_ = need_is_none.pop()
    if not type_ in done_zero:
        c = generate_zero_method(type_)
        if c is not None:
            code += c
    done_zero.add(type_)

with open(OUT, "w") as f:
    f.write(code)
