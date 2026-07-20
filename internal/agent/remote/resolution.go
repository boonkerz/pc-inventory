package remote

import "math"

// pickResolution wählt aus den angebotenen Modi den, dessen Seitenverhältnis
// (vorrangig) und Größe am besten zur Zielgröße (tw,th) passt. So verschwinden die
// schwarzen Ränder im Viewer, ohne riskante eigene Modelines anzulegen. Plattform-
// neutral und damit unabhängig von Win32 testbar.
func pickResolution(modes [][2]int, tw, th int) (int, int, bool) {
	if tw <= 0 || th <= 0 {
		return 0, 0, false
	}
	targetAspect := float64(tw) / float64(th)
	bestW, bestH := 0, 0
	bestCost := math.MaxFloat64
	for _, m := range modes {
		mw, mh := m[0], m[1]
		if mw < 640 || mh < 480 {
			continue
		}
		aspErr := math.Abs(float64(mw)/float64(mh) - targetAspect)
		sizeErr := math.Abs(float64(mw-tw)) + math.Abs(float64(mh-th))
		cost := aspErr*100000 + sizeErr // Seitenverhältnis dominiert
		if cost < bestCost {
			bestCost = cost
			bestW, bestH = mw, mh
		}
	}
	if bestW == 0 {
		return 0, 0, false
	}
	return bestW, bestH, true
}
