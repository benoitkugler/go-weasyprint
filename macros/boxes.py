import re
import subprocess
import inspect

import source_box

OUT = "structure/stubs.go"

RE_CLASS = re.compile(r"class (\S+)\(")
RE_METHOD = re.compile(r"    def (\S+)\((.*)\):")
RE_ATTRIBUTES = re.compile(r"^    (\S+) = (\S+)")
RE_GO_FIELDS = re.compile(r"\t(\w+,? )+")


def class_by_name(class_name):
    return getattr(source_box, class_name)


def expand_supers(class_name):
    level = set(c.__name__ for c in class_by_name(
        class_name).__bases__ if c.__name__ != "object")
    out = set(level)
    for parent in level:
        out = out.union(expand_supers(parent))
    return out


def genere_interface(class_name, parents) -> str:
    parents_inter = "\n".join(
        f'is{parent}()' for parent in parents)
    i = f"""
        type Instance{class_name} interface {{
            is{class_name} ()
            {parents_inter}
        }}"""
    return i


def genere_concrete_type(class_name, new_fields) -> str:
    fs = "\n".join(to_camel_case(
        f[0]) + " " + infer_type(f[1]) + "// " + f[1] for f in new_fields)

    if issubclass(class_by_name(class_name), source_box.TableBox):
        fs = "TableFields  \n" + fs
    return f"""
    type {class_name} struct {{
        BoxFields
        Instance{class_name}

        {fs}
    }}
    """


def to_kebab_case(name):
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def to_camel_case(name):
    output = ''.join(x for x in name.title() if x.isalnum())
    return output[0].lower() + output[1:]


def to_go(lit):
    return lit.replace("'", '"').replace("True", "true").replace("False", "false")


def infer_type(value: str):
    if value in ("true", "false"):
        return "bool"
    elif value in ("0", "1"):
        return "int"
    elif value.startswith('"'):
        return "string"
    elif value == "None":
        return "interface{}"
    raise ValueError(value)


TYPES_VARIABLES = {
    "dx": "float32",
    "dy": "float32",
    "ignore_floats": "bool",
    "side": "string",
    "start": "bool",
    "end": "bool",
    "is_start": "bool",
    "is_end": "bool",
    "new_children": "[]Box",
    "text": "string"
}

TYPES_RETURNS = {
    "all_children": "[]Box",
    "copy_with_children": "Box",
    "deepcopy": "Box",
    "descendants": "[]Box",
    "get_wrapped_table": "(Box, error)",
    "page_values": "(pr.Page, pr.Page)",
    "hit_area": "(x float32, y float32, w float32, h float32)",
    "get_cells": "[]Box",
}


def has_subclass(class_name):
    return len(class_by_name(class_name).__subclasses__()) > 0


def generate_own_method(name: str, class_name: str, args: list):
    comment = " ".join(arg for arg in args if "=" in arg)
    args_no_selfs = []
    args_no_types = []
    for arg in args[1:]:
        arg = arg.split("=")[0]
        type_ = TYPES_VARIABLES[arg]
        arg = to_camel_case(arg)
        args_no_selfs.append(arg + " " + type_)
        args_no_types.append(arg)
    args_no_self = ", ".join(args_no_selfs)
    args_no_type = ", ".join(args_no_types)

    ret = TYPES_RETURNS.get(name, "")

    name = to_camel_case(name)
    need_in_subclass = has_subclass(class_name)
    header = ""
    if need_in_subclass:
        header_func = f"{class_name}{name[0].title() + name[1:]}"
        header = f"func {header_func} (b Instance{class_name}, {args_no_self})  {ret} {{}}\n"
        body = f"{header_func}(b, {args_no_type})"
    else:
        body = "//TODO" + ":"

    signature = f"{name} ({args_no_self}) {ret}"
    has_return = ret != ""
    out = f"""
    // {comment}
    func (b *{class_name}) {signature} {{
        {"return" if has_return else ""} {body}
    }}"""

    if need_in_subclass:
        outer_methods[(class_name, name)] = (body, signature, has_return)
        return out, header
    return "", out


def generate_inherited_methods(class_name: str) -> str:
    cls = class_by_name(class_name)
    out = ""
    for name, func in inspect.getmembers(cls, inspect.isfunction):
        owner = func.__qualname__.split(".")[0]
        # already implemented by generate_own_methods
        if owner == class_name or owner == "Box" or name == "__repr__":
            continue

        if name == "__init__":
            init_fields = "\n".join(
                f"out.{field} = {value}" for field, value in classes[class_name])
            out += f"""
            func New{class_name}(elementTag string, style pr.Properties, children []Box) {class_name} {{
                fields := newBoxFields(elementTag, style, children)
                out := {class_name}{{BoxFields: fields}}
                {init_fields}
                return out
            }}
            """
            continue

        call, sign, has_return = outer_methods[(owner, to_camel_case(name))]

        out += f"""
        func (b {class_name}) {sign} {{
            { "return" if has_return else ""} {call}
        }}
        """
    return out


KNOW_FIELDS = set()
with open("structure/boxes_impl.go") as f:
    for line in f.readlines()[16:67]:
        match = RE_GO_FIELDS.search(line)
        if match:
            KNOW_FIELDS = KNOW_FIELDS.union(to_kebab_case(
                n) for n in match.group(0).strip().split(", "))


classes = {}
outer_methods = {}

with open("macros/source_box.py") as f:
    code = """
    package structure

    // autogenerated from source_box.py
    """
    headers = ""
    class_name = ""
    for line in f.readlines():
        match = RE_METHOD.search(line)
        if match and class_name != "Box":
            name = match.group(1)
            if name not in ("__init__", "anonymous_from"):
                args = match.group(2).split(", ")
                out, header = generate_own_method(
                    match.group(1), class_name, args)
                code += out
                headers += header
            elif name == "__init__":
                headers += f"""
                func New{class_name}(elementTag string, style pr.Propreties) {class_name} {{
                    // TODO:
                }}
                """

        match = RE_ATTRIBUTES.search(line)
        if match:
            field, value = match.group(1), match.group(2)
            value = to_go(value)
            if not field in KNOW_FIELDS and field not in ("proper_parents", "ascii_to_wide"):
                classes[class_name].append((field, value))

        match = RE_CLASS.search(line)
        if match:
            class_name = match.group(1)
            parents = expand_supers(class_name)
            classes[class_name] = []
            code += genere_interface(class_name, parents)

for class_name, fields in classes.items():
    code += genere_concrete_type(class_name, fields)
    code += generate_inherited_methods(class_name)

with open(OUT, "w") as f:
    f.write(code)

subprocess.run(["goimports", "-w", OUT])

print(headers)
