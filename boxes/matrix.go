package boxes

// Matrix encode a (2D) linear transformation (Y = AX + B)

type Matrix interface {
	Copy() Matrix

	// Data() method returns a canonical version
	// given by the following convention.
	// Assuming xx, yx, xy, yy, x0, y0 == [6]float64,
	// the transformation of a point (x,y) is given by:
	// 	x_new = xx * x + xy * y + x0
	// 	y_new = yx * x + yy * y + y0
	// which is equivalent to :
	// A = 	| xx xy |  B = 	| x0 |
	//		| yx yy | 		| y0 |
	Data() [6]float64

	// Returns other * self
	LeftMultiply(other [6]float64) Matrix
	// Returns self * other
	RightMultiply(other [6]float64) Matrix

	// Invert modify the matrix in place. Return an error
	// if the transformation is not bijective.
	Invert() error

	// Transforms the point `(x, y)` by this matrix
	TransformPoint(x, y float64) (outX, outY float64)

	// Applies a translation by `tx`, `ty`
	// to the transformation in this matrix.
	//
	// The effect of the new transformation is to
	// first translate the coordinates by `tx` and `ty`,
	// then apply the original transformation to the coordinates.
	//
	// 	This changes the matrix in-place.
	Translate(tx, ty float64)

	// Applies scaling by `sx`, `sy`
	// to the transformation in this matrix.
	//
	// The effect of the new transformation is to
	// first scale the coordinates by `sx` and `sy`,
	// then apply the original transformation to the coordinates.
	//
	// This changes the matrix in-place.
	Scale(sx, sy float64)

	// Applies a rotation by `radians`
	// to the transformation in this matrix.
	//
	// The effect of the new transformation is to
	// first rotate the coordinates by `radians`,
	// then apply the original transformation to the coordinates.
	//
	// 	This changes the matrix in-place.
	//
	// `radians` is the angle of rotation, in radians.
	// 	The direction of rotation is defined such that positive angles
	// 	rotate in the direction from the positive X axis
	// 	toward the positive Y axis.
	// 	With the default axis orientation of cairo,
	// 	positive angles rotate in a clockwise direction.
	Rotate(radians float64)
}
