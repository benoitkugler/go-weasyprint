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


IN = [('p', 'Block', [('p', 'Line', [('p', 'Text', 'start '), ('i', 'Inline', [('b', 'Inline', [('b', 'Text', 'bi1')]), ('i', 'Text', ' '),
                                                                               ('b', 'Inline', [('b', 'Text', 'bi2')])]), ('p', 'Text', ' '), ('b', 'Inline', [('b', 'Text', 'b1')]), ('p', 'Text', ' end')])])]

pyperclip.copy(to_go(IN))
print("Copied in clipboard.")
