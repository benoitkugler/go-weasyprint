package version

import (
	"fmt"
)

const (
	Version = "0.50"
)

// Used for "User-Agent" in HTTP and "Creator" in PDF
var VersionString = fmt.Sprintf("Go-WeasyPrint %s (http://weasyprint.org/)", Version)
