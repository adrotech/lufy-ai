// release-contract emite metadata estática para los checksums del release.
package main

import (
	"fmt"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/upgrade"
)

func main() { fmt.Print(upgrade.RuntimeContractHeader()) }
