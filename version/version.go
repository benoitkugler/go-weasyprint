package version

import "fmt"

const (
	Version = "0.42.3"
)

// Used for "User-Agent" in HTTP and "Creator" in PDF
var VersionString = fmt.Sprintf("Go-WeasyPrint %s (http://weasyprint.org/)", Version)
