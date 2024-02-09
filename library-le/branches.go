package libraryle

import "strings"

var branchNames = map[string]int{
	"stadtbibliothek": 0,
	"plagwitz":        20,
	"wiederitzsch":    21,
	"böhlitz":         22,
	"lützschena":      23,
	"holzhausen":      25,
	"südvorstadt":     30,
	"gohlis":          41,
	"volkmarsdorf":    50,
	"schönefeld":      51,
	"paunsdorf":       60,
	"reudnitz":        61,
	"mockau":          70,
	"grünau-mitte":    82,
	"grünau-nord":     83,
	"grünau-süd":      84,
}

// Liefert den den BranchCode sofern existent.
func GetBranchCode(branchNameQuery string) (int, bool) {
	branchCode, present := branchNames[strings.ToLower(branchNameQuery)]
	return branchCode, present
}
