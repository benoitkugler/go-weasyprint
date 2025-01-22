# go-weasyprint

[![Go Reference](https://pkg.go.dev/badge/github.com/benoitkugler/go-weasyprint.svg)](https://pkg.go.dev/github.com/benoitkugler/go-weasyprint)
[![Build Status](https://github.com/benoitkugler/go-weasyprint/actions/workflows/build.yml/badge.svg)](https://github.com/benoitkugler/go-weasyprint/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/benoitkugler/go-weasyprint)](https://goreportcard.com/report/github.com/benoitkugler/go-weasyprint)
[![Latest release](https://img.shields.io/github/release/benoitkugler/go-weasyprint.svg)](https://github.com/benoitkugler/go-weasyprint/releases)

Golang port of [Weasyprint](https://github.com/Kozea/WeasyPrint) python Html to Pdf library.

This project is still in alpha state; you should try it carefully before using it in production. 

## Outline

This package converts an HTML document (with its associated CSS files) to a PDF file.
The heavy lifting is actually delegated to [webrender](https://github.com/benoitkugler/webrender), but this package implements a backend for PDF files, relying on [benoitkugler/pdf](https://github.com/benoitkugler/pdf).
