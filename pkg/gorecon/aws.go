package gorecon

import (
	"github.com/D3Ext/go-recon/core"
)

// this function returns all defined permutations for
// S3 buckets name generation
func GetAllPerms() []string {
	return core.GetAllPerms()
}

// this function returns more or less permutations based on given level
// 1 returns less permutations than 6 (1 lower, 5 higher)
func GetPerms(level int) []string {
	return core.GetPerms(level)
}
