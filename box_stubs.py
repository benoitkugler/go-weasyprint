import re

PACKAGE = "structure2"
SOURCE = f"{PACKAGE}/boxes.go"
TARGET = f"{PACKAGE}/generated.go"

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
TEMPLATE_COPY = """
    func (b  {name}) Copy() Box {{ return &b }}
    """

TEMPLATE_PARENT = """parent := New{parent}({args_no_type})
    out.{parent} = *parent"""

TEMPLATE_ANONYMOUS = """func {name}AnonymousFrom(parent Box, {arg}) *{name} {{
	return New{name}(parent.Box().elementTag, parent.Box().style.InheritFrom(), {arg_no_type})
}}"""

TEMPLATE_TYPE = """
func (t type{name}) AnonymousFrom(parent Box, children []Box) Box {{
	return {name}AnonymousFrom(parent, children)
}}
"""
TEMPLATE_IS_INSTANCE = """func (t type{name}) IsInstance(box Box) bool {{
	_, is := box.(*{name})
	return is
}}
"""

ANONYMOUS_FROM = ("TableRowBox", "TableRowGroupBox", "TableColumnBox", "TableColumnGroupBox",
                  "TableBox", "TableCellBox", "InlineTableBox")
IS_INSTANCE = ("TableRowBox", "TableRowGroupBox", "TableColumnBox", "TableColumnGroupBox",
               "TableCellBox", "InlineTableBox", "TextBox")
TYPES = set(ANONYMOUS_FROM + IS_INSTANCE)


def parse_constructor(s: str):
    args = s.split(", ")
    parsed_args = list(arg.split(" ") for arg in args)
    return parsed_args


def render_init(l: list):
    return "".join([f"{field}: {value},\n" for field, value in l])


with open(SOURCE) as f:
    lines = f.readlines()
    args, init_fiels, is_in = None, [], False
    var, types_ = "", ""

    full_code = f"""package {PACKAGE}

    // autogenerated from boxes.go

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
                parsed_args[-1][0] = "nil"

            args_no_type = ", ".join(t[0] for t in parsed_args)
            init = render_init(init_fiels)
            if parent is not None:
                parent = TEMPLATE_PARENT.format(
                    parent=parent, args_no_type=args_no_type)
            else:
                parent = ""
            code = TEMPLATE.format(name=name, args=args,
                                   args_no_type=args_no_type, parent=parent, init=init)

            if name not in ("TextBox", "PageBox", "MarginBox", "LineBox"):
                full_code += code

            if name in ANONYMOUS_FROM:
                types_ += TEMPLATE_TYPE.format(name=name) + "\n"

            if name in IS_INSTANCE:
                types_ += TEMPLATE_IS_INSTANCE.format(name=name)

            if name in TYPES:
                var += f"Type{name} BoxType = type{name}{{}} \n"
                types_ += f"type type{name} struct{{}}"

            code = TEMPLATE_ANONYMOUS.format(
                name=name, arg=" ".join(parsed_args[-1]), arg_no_type=parsed_args[-1][0])
            if name not in ("PageBox", "MarginBox"):
                full_code += code

            full_code += TEMPLATE_COPY.format(name=name)+"\n"

            args = None
            is_in = False

    full_code += f"var ( \n {var} \n ) \n" + types_

with open(TARGET, "w") as f:
    f.write(full_code)
