def to_go(boxes: list) -> str:
    code = "[]serializedBox{"
    for box in boxes:
        tag = box[0]
        type_ = f"Type{box[1]}Box"
        if isinstance(box[2], str):

            content = 'bc{{text: "{0}"}}'.format(repr(box[2])[1:-1])
        else:
            children = to_go(box[2])
            content = f"bc{{c: {children}}}"
        code += f"""{{"{tag}", {type_}, {content}}},\n"""
    code += "}"
    return code


IN = [
    ("body", "Line", [
        ("p", "InlineBlock", [
            ("p", "Block", [
                ("p", "Line", [
                    ("p", "Text", "Lorem "),
                    ("em", "Inline", [
                        ("em", "Text", "ipsum "),
                        ("strong", "Inline", [
                            ("strong", "Text", "dolor ")])])])]),
            ("span", "Block", [
                ("span", "Line", [
                    ("span", "Text", "sit")])]),
            ("p", "Block", [
                ("p", "Line", [
                    ("em", "Inline", [
                        ("strong", "Inline", [
                            # Whitespace processing ! done yet.
                            ("strong", "Text", "\n      ")])])])]),
            ("span", "Block", [
                ("span", "Line", [
                    ("span", "Text", "amet,")])]),

            ("p", "Block", [
                ("p", "Line", [
                    ("em", "Inline", [
                        ("strong", "Inline", [])])])]),
            ("span", "Block", [
                ("span", "Block", [
                    ("span", "Line", [
                        ("em", "Inline", [
                            ("em", "Text", "conse")])])]),
                ("i", "Block", []),
                ("span", "Block", [
                    ("span", "Line", [
                        ("em", "Inline", [])])])]),
            ("p", "Block", [
                ("p", "Line", [
                    ("em", "Inline", [])])])])])]

print(to_go(IN))
