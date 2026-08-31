package server

// Location codes known at the time of writing.
//
// These are conveniences for readable call sites, not an authoritative
// list. Locations are added and retired, and what a given account can
// provision into depends on the product as well as the location — the
// ProductTypes field on each [Location] row says which product
// families a location carries. Call [Client.ListLocations] when the
// answer has to be current, and treat these constants as hints.
//
// Note that distinct codes can share a data centre label
// ([LocationAKLCity] and [LocationWINCity] are both AKL01, differing
// in which product families they carry), so match on the code and
// never on the label.
const (
	// LocationAKLCity is NZ - AKL01.
	LocationAKLCity = "AKLCITY"

	// LocationAKLNCT is NZ - AKL02.
	LocationAKLNCT = "AKLNCT"

	// LocationWINCity is the Windows-capable AKL01 pool.
	LocationWINCity = "WINCITY"

	// LocationWINNCT is the Windows-capable AKL02 pool.
	LocationWINNCT = "WINNCT"

	// LocationLINSYD1 is AU - SYD01 (Linux).
	LocationLINSYD1 = "LINSYD1"

	// LocationWINSYD1 is AU - SYD01 (Windows).
	LocationWINSYD1 = "WINSYD1"

	// LocationFRA1 is DE - FRA1.
	LocationFRA1 = "FRA1"

	// LocationUSCAL1 is US - CAL1.
	LocationUSCAL1 = "USCAL1"
)
