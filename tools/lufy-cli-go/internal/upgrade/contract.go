package upgrade

import (
	"bytes"
	"fmt"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

// RuntimeContractHeader identifica el único modelo que admite esta compilación.
// Se genera en checksums; no ejecutamos el binario descargado para inspeccionarlo.
func RuntimeContractHeader() string {
	return fmt.Sprintf("# lufy-runtime: execution=%s config=%d state=%d backup=%d\n", domain.ExecutionInline, projectconfig.SchemaVersion, state.SchemaVersion, backup.SchemaVersion)
}
func validateRuntimeContract(checksums []byte) error {
	first, _, _ := bytes.Cut(checksums, []byte("\n"))
	if !bytes.Equal(first, bytes.TrimSuffix([]byte(RuntimeContractHeader()), []byte("\n"))) {
		return fmt.Errorf("release no declara el modelo inline actual; se conservó el binario instalado")
	}
	return nil
}
