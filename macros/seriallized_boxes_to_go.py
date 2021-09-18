import pyperclip


def to_go(boxes: list) -> str:
    code = "[]serBox{"
    for box in boxes:
        tag = box[0]
        type_ = f"{box[1]}BoxT"
        if isinstance(box[2], str):

            content = 'bc{{text: "{0}"}}'.format(repr(box[2])[1:-1])
        else:
            children = to_go(box[2])
            content = f"bc{{c: {children}}}"
        code += f"""{{"{tag}", {type_}, {content}}},\n"""
    code += "}"
    return code


IN = [
    ("p", "Block", [
        ("p", "Line", [
            ("p::before", "Inline", [
                ("p::before", "Text", "a"),
                ("p::before", "InlineReplaced", "<replaced>"),
                ("p::before", "Text", "b")]),
            ("p", "Text", "c")])])]
pyperclip.copy(to_go(IN))
print("Copied in clipboard.")
