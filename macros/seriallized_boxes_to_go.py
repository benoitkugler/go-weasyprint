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
    ("x-table", "Block", [
        ("x-caption", "TableCaption", [
                ("x-caption", "Line", [
                    ("x-caption", "Text", "top caption")])]),
        ("x-table", "Table", [
            ("x-table", "TableColumnGroup", [
                    ("x-col", "TableColumn", [])]),
            ("x-thead", "TableRowGroup", [
                ("x-thead", "TableRow", [
                        ("x-th", "TableCell", [])])]),
            ("x-table", "TableRowGroup", [
                ("x-tr", "TableRow", [
                        ("x-th", "TableCell", [
                            ("x-th", "Line", [
                                ("x-th", "Text", "foo")])]),
                        ("x-th", "TableCell", [
                            ("x-th", "Line", [
                                ("x-th", "Text", "bar")])])])]),
            ("x-thead", "TableRowGroup", []),
            ("x-table", "TableRowGroup", [
                ("x-tr", "TableRow", [
                        ("x-td", "TableCell", [
                            ("x-td", "Line", [
                                ("x-td", "Text", "baz")])])])]),
            ("x-tfoot", "TableRowGroup", [])]),
        ("x-caption", "TableCaption", [])])]

print(to_go(IN))
