# go-weasyprint

Golang port of [Weasyprint](https://github.com/Kozea/WeasyPrint) python Html to Pdf library.

This is an **ongoing work**, not production ready just yet.

## Outline

This module implements a static HTML renderer, which works by :

- parsing the HTML input and fetching CSS files, and cascading the styles. This is implemented in the `tree` package

- building a tree of boxes from the HTML structure (package `boxes`)

- laying out this tree, that is attributing position and dimensions to the boxes, and performing line, paragraph and page breaks (package `layout`)

- drawing the laid out tree to an output. Contrary to the Python library, this step is here performed on an abtract output, which must implement the `backend.Output` interface. This means than the core layout logic could easily be reused for other purposes.

- the package `pdf` implements one output for PDF files, building on `benoitkugler/pdf/model`
