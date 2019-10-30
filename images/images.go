package images

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

type Image interface {
	isImage()
	GetIntrinsicSize(imageResolution, fontSize pr.Value) (width, height pr.MaybeFloat)
	IntrinsicRatio() pr.MaybeFloat
}
