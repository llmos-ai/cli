package init

import "github.com/llmos-ai/cli/pkg/logserver"

func init() {
	go logserver.StartServerWithDefaults()
}
