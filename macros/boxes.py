import inspect
import re
import subprocess

import source_box

OUT = "boxes/stubs.go"

RE_CLASS = re.compile(r"class (\S+)\(")
RE_METHOD = re.compile(r"    def (\S+)\((.*)\):")
RE_ATTRIBUTES = re.compile(r"^    (\S+) = (.+)\n")
RE_GO_FIELDS = re.compile(r"\t(\w+,? )+")

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
    "text": "string",
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
    "__repr__": "string",
}

DEFAULT_CONSTRUCTOR = [("elementTag", "string"),
                       ("style", "pr.Properties"), ("children", "[]Box")]
SPECIALS_CONSTRUCTORS = {
    "TextBox": [("elementTag", "string"), ("style", "pr.Properties"), ("text", "string")],
    "ReplacedBox": [("elementTag", "string"), ("style", "pr.Properties"), ("replacement", "images.Image")],
    "BlockReplacedBox": [("elementTag", "string"), ("style", "pr.Properties"), ("replacement", "images.Image")],
    "InlineReplacedBox": [("elementTag", "string"), ("style", "pr.Properties"), ("replacement", "images.Image")],
    "PageBox": [("pageType", "pr.PageElement"), ("style", "pr.Properties")],
    "MarginBox": [("atKeyword", "string"), ("style", "pr.Properties")],
}

ABSTRACT_TYPES = {"BlockLevelBox", "InlineLevelBox", "BlockContainerBox",
                  "ParentBox", "AtomicInlineLevelBox", "FlexContainerBox"}


def infer_type(value: str):
    if value in ("true", "false"):
        return "bool"
    elif value in ("0", "1"):
        return "int"
    elif value == '"PageType"':
        return "utils.PageElement"
    elif value == '"Image"':
        return "pr.Image"
    elif value.startswith('"'):
        return "string"
    elif value == "None":
        return "interface{}"
    raise ValueError(value)


def class_by_name(class_name):
    return getattr(source_box, class_name)


def expand_supers(class_name):
    level = set(c.__name__ for c in class_by_name(
        class_name).__bases__ if c.__name__ != "object")
    out = set(level)
    for parent in level:
        out = out.union(expand_supers(parent))
    return out


common_methods = {"translate", "removeDecoration", "deepcopy", "descendants"}


def genere_interface(class_name) -> str:
    """ Creates interfaces and trivial implementations """
    parents = expand_supers(class_name)

    type_methods = [f"is{class_name} ()"] + \
        [f'is{parent}()' for parent in parents if parent != "Box"]

    # normal_methods = []
    # for owner, name, _ in expand_methods(class_name):
    #     if name != "__init__" and name != "__repr__" and owner != "Box" and not name in common_methods:
    #         _, sign, _ = outer_methods[(owner, to_camel_case(name))]
    #         normal_methods.append(sign)

    s = "\n".join(type_methods)

    comment = format_comment(inspect.getdoc(class_by_name(class_name)))

    i = f"""
        {comment} 
        type instance{class_name} interface {{
            {s}
        }}
        """

    if class_name in ABSTRACT_TYPES:
        return i

    i += "\n".join(f"func({class_name}) {meth} {{}}" for meth in type_methods)
    i += f"""
    func (b *{class_name}) Box() *BoxFields {{ return &b.BoxFields }}

    // Copy is a shallow copy
    func (b {class_name}) Copy() Box {{ return &b }}
    """

    if not class_name in ("MarginBox", "PageBox"):
        i += f"""
        func (b {class_name}) String() string {{
            return fmt.Sprintf("<{class_name} %s>", b.BoxFields.elementTag)
        }}
        """
    return i


# def genere_concrete_type(class_name, new_fields) -> str:
#
#     fs = "\n".join(to_camel_case(
#         f[0]) + " " + infer_type(f[1]) + "// " + f[1] for f in new_fields)

#     if issubclass(class_by_name(class_name), source_box.TableBox):
#         fs = "TableFields  \n" + fs
#     return f"""
#     {comment}
#     type {class_name} struct {{
#         BoxFields

#         {fs}
#     }}
#     """


def to_kebab_case(name):
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()


def to_camel_case(name):
    output = ''.join(x for x in name.title() if x.isalnum())
    return output[0].lower() + output[1:]


def to_go(lit):
    return lit.replace("'", '"').replace("True", "true").replace("False", "false")


def has_subclass(class_name):
    return len(class_by_name(class_name).__subclasses__()) > 0


# def generate_own_method(name: str, class_name: str, args: list):
#     comment = " ".join(arg for arg in args if "=" in arg)
#     args_no_selfs = []
#     args_no_types = []
#     for arg in args[1:]:
#         arg = arg.split("=")[0]
#         type_ = TYPES_VARIABLES[arg]
#         arg = to_camel_case(arg)
#         args_no_selfs.append(arg + " " + type_)
#         args_no_types.append(arg)
#     args_no_self = ", ".join(args_no_selfs)
#     args_no_type = ", ".join(args_no_types)

#     ret = TYPES_RETURNS.get(name, "")

#     name = to_camel_case(name)
#     if name == "repr":
#         name = "String"
#     need_in_subclass = has_subclass(class_name)
#     header = ""
#     if need_in_subclass:
#         header_func = f"{class_name.lower()}{name[0].title() + name[1:]}"
#         header = f"func {header_func} (b instance{class_name}, {args_no_self})  {ret} {{}}\n"
#         body = f"{header_func}(b, {args_no_type})"
#     else:
#         body = "//TODO" + ":"

#     signature = f"{name} ({args_no_self}) {ret}"
#     has_return = ret != ""
#     out = f"""
#     // {comment}
#     func(b * {class_name}) {signature} {{
#         {"return" if has_return else ""} {body}
#     }}"""

#     outer_methods[(class_name, name)] = (body, signature, has_return)
#     if need_in_subclass:
#         return out, header
#     return "", out


def expand_methods(class_name: str) -> list:
    """ Returns the list of all methods(including inherited)"""
    cls = class_by_name(class_name)
    out = []
    for name, func in inspect.getmembers(cls, inspect.isfunction):
        comment = inspect.getdoc(func)
        owner = func.__qualname__.split(".")[0]
        out.append((owner, name, comment))
    return out


def format_comment(python_comment: str):
    python_comment = python_comment or ""
    out = python_comment.replace("\n\n", "\n").replace('"""', "")
    return "\n".join("// " + line for line in out.split("\n"))


def generate_constructor(class_name):
    args = SPECIALS_CONSTRUCTORS.get(class_name, DEFAULT_CONSTRUCTOR)
    fields_values = dict(classes[class_name])
    with_children = args[-1][0] == 'children'
    if not with_children:
        fields_values[args[-1][0]] = args[-1][0]
    init_fields = "\n".join(
        f"out.{to_camel_case(field)} = {value}" for field, value in fields_values.items())
    return f"""
    func New{class_name}({", ".join(arg[0] + " " + arg[1] for arg in args)}) {class_name} {{
        out := {class_name}{{}} // TODO:
        {init_fields}
        return out
    }}
    """


# def generate_inherited_methods(class_name: str) -> str:
#     out = ""
#     for owner, name, comment in expand_methods(class_name):
#         if name == "__repr__" and owner != class_name:
#             out += f"""
#             func (b {class_name}) String() string {{
#                 return fmt.Sprintf("<{class_name} %s>", b.elementTag)
#             }}
#             """

#         # already implemented by generate_own_methods
#         if owner == class_name or owner == "Box":
#             continue

#         call, sign, has_return = outer_methods[(owner, to_camel_case(name))]

#         out += f"""
#         {format_comment(comment)}
#         func (b {class_name}) {sign} {{
#             { "return" if has_return else ""} {call}
#         }}
#         """
#     return out


def generate_type_objects(class_name: str):
    if class_name in ABSTRACT_TYPES:
        return f"""
        func Is{class_name}(box Box) bool {{
        	_, is := box.(instance{class_name})
        	return is
        }}
        """, None, None

    own_anonymous_from = False
    for name, func in inspect.getmembers(class_by_name(class_name), inspect.ismethod):
        if name == "anonymous_from" and func.__qualname__.split(".")[0] == class_name:
            own_anonymous_from = True
            break

    args = SPECIALS_CONSTRUCTORS.get(class_name, None)
    is_normal = args == None
    if is_normal:
        args = DEFAULT_CONSTRUCTOR
    body = "// TODO" + ":"
    if not own_anonymous_from:
        body = f"""style := tree.ComputedFromCascaded(nil, nil, parent.Box().Style, nil, "", "", nil)
            out := New{class_name}(parent.Box().elementTag, style, {args[-1][0]})
            return &out
            """

    out = ""

    fn = f"""
        func {class_name}AnonymousFrom(parent Box, {args[-1][0] + " " + args[-1][1]}) *{class_name} {{
            {body}
        }}
        """
    header = ""
    if own_anonymous_from:
        header = fn
    elif class_name not in ("MarginBox", "PageBox"):
        out += fn
    var_type = None
    if is_normal:  # we cant include others for signature reason
        out += f"""
        func (t type{class_name})IsInstance(box Box) bool {{
        	_, is := box.(instance{class_name})
        	return is
        }}

        func (b *{class_name}) Type() BoxType {{
            return Type{class_name}
        }}
        type type{class_name} struct{{}}

        func (t type{class_name}) AnonymousFrom(parent Box, children []Box) Box {{
            return {class_name}AnonymousFrom(parent, children)
        }}
        """
        var_type = f"Type{class_name} BoxType = type{class_name}{{}}"
    else:
        out += f"""
        func Is{class_name}(box Box) bool {{
        	_, is := box.(instance{class_name})
        	return is
        }}
        """
    return out, var_type, header


def genere_proper_parents(class_name, value):
    parents = value[1:-1].split(",")
    return f"""
    func ({class_name}) IsProperChild(parent Box) bool {{
        switch parent.(type) {{
        case {','.join("*" + p for p in parents)}:
            return true
        default:
            return false
        }}
    }}
    """


KNOW_FIELDS = set()
with open("boxes/boxes.go") as f:
    for line in f.readlines()[83:135]:
        match = RE_GO_FIELDS.search(line)
        if match:
            KNOW_FIELDS = KNOW_FIELDS.union(to_kebab_case(
                n) for n in match.group(0).strip().split(", "))


classes = {}
outer_methods = {}

with open("macros/source_box.py") as f:
    code = """
    package boxes

    // autogenerated from source_box.py
    """
    headers = ""
    class_name = ""
    for line in f.readlines():
        # match = RE_METHOD.search(line)
        # if match and class_name != "Box":
            # name = match.group(1)

            # if name not in ("__init__", "anonymous_from"):
            #     args = match.group(2).split(", ")
            #     out, header = generate_own_method(
            #         name, class_name, args)
            #     code += out
            #     headers += header
            # elif name == "__init__" and class_name != "ParentBox":
            #     headers += f"""
            #     func New{class_name}(elementTag string, style pr.Properties) {class_name} {{
            #         // TODO:
            #     }}
                # """

        match = RE_ATTRIBUTES.search(line)
        if match:
            field, value = match.group(1), match.group(2)
            if field == "proper_parents":
                code += genere_proper_parents(class_name, value)

            value = to_go(value)
            if not field in KNOW_FIELDS and field not in ("proper_parents", "ascii_to_wide"):
                classes[class_name].append((field, value))

        match = RE_CLASS.search(line)
        if match:
            class_name = match.group(1)
            classes[class_name] = []

var_types = []
for class_name, fields in classes.items():
    if class_name != "Box":
        # code += genere_concrete_type(class_name, fields)
        # code += generate_inherited_methods(class_name)
        code += genere_interface(class_name)
        decl, var, header = generate_type_objects(class_name)
        if header:
            headers += header
        headers += generate_constructor(class_name)
        var_types.append(var)
        code += decl

var_type = "\n".join(v for v in var_types if v)
code += f"""
var(
    {var_type}
)
"""


with open(OUT, "w") as f:
    f.write(code)

subprocess.run(["goimports", "-w", OUT])

print("""
package structure


""" + headers)
