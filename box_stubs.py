import re

PACKAGE = "structure2"
SOURCE = f"{PACKAGE}/boxes.go"
TARGET = f"{PACKAGE}/boxes_def.go"

RE_TYPE = re.compile(r"^type (\w+)Box struct ")
RE_FIELDS = re.compile(r"(\w+)\s+\w+ // init:(\S+)")
RE_CONSTRUCTOR = re.compile(r"// constructor:(.+)")

DEFAULT_ARGS = "elementTag string, style css.StyleDict, children []Box"

TEMPLATE = """func New{name}({args}) *{name} {{
    out := {name}{{
    {init}}}
    {parent}
    return &out
}}
"""

TEMPLATE_PARENT = """parent := New{parent}({args_no_type})
    out.{parent} = *parent"""


def parse_constructor(s: str):
    args = s.split(", ")
    parsed_args = list(arg.split(" ")[0] for arg in args)
    return parsed_args


def render_init(l: list):
    return "".join([f"{field}: {value},\n" for field, value in l])


with open(SOURCE) as f:
    lines = f.readlines()
    args, init_fiels, is_in = None, [], False
    full_code = f"""package {PACKAGE}

    """
    full_code += TEMPLATE.format(name="BoxFields",
                                 args=DEFAULT_ARGS, init=render_init([
                                     ("elementTag", "elementTag"),
                                     ("style", "style"),
                                     ("children", "children"),
                                 ]), parent="")
    for i, line in enumerate(lines):
        match = RE_TYPE.search(line)
        if match:
            name = match.group(1) + "Box"
            parent = lines[i+1].strip()
            if parent == "}" or len(parent.split(" ")) > 1:
                parent = None
            is_in = True
            init_fiels = []

        match = RE_FIELDS.search(line)
        if match:
            field, value = match.group(1), match.group(2)
            init_fiels.append((field, value))

        match = RE_CONSTRUCTOR.search(line)
        if match:
            args = match.group(1)

        if is_in and line == "}\n":
            args = args if args else DEFAULT_ARGS
            parsed_args = parse_constructor(args)

            if name == "ReplacedBox":
                init_fiels.append(("replacement", "replacement"))
                parsed_args[-1] = "nil"

            args_no_type = ", ".join(parsed_args)
            init = render_init(init_fiels)
            if parent is not None:
                parent = TEMPLATE_PARENT.format(
                    parent=parent, args_no_type=args_no_type)
            else:
                parent = ""
            code = TEMPLATE.format(name=name, args=args,
                                   args_no_type=args_no_type, parent=parent, init=init)
            full_code += code
            args = None
            is_in = False

with open(TARGET, "w") as f:
    f.write(full_code)
