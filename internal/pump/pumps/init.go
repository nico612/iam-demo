package pumps

var availablePumps map[string]Pump

// nolint: gochecknoinits
func init() {
	availablePumps = make(map[string]Pump)

	availablePumps["mongo"] = &MongoPump{}
}
