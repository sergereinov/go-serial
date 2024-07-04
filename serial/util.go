// ------------------------------------------
// Modified by (c) 2024 Serge Reinov.
//   Moved to a separate file.
//
// Licensed under the Apache License, Version 2.0.
// ------------------------------------------

package serial

import "math"

// Rounds a float to the nearest integer.
func round(f float64) float64 {
	return math.Floor(f + 0.5)
}
