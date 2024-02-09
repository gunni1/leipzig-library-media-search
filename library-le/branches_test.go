package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestGetBranchCodeSuccess(t *testing.T) {
	branchNameQuery := "gohlis"
	result, present := GetBranchCode(branchNameQuery)
	True(t, present)
	Equal(t, result, 41)
}

func TestGetBranchCodeUnknown(t *testing.T) {
	branchNameQuery := "blubb"
	_, present := GetBranchCode(branchNameQuery)
	False(t, present)
}
