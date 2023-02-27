package streamer

import (
	"encoding/base64"
	"fmt"
	"strings"

	"keepair/pkg/common"
)

const Seperator = ","

func EncodeMessage(entry common.Entry) (string, error) {
	k := entry.Key
	if strings.Contains(k, Seperator) {
		return "", fmt.Errorf("key cannot contain '%s' character", Seperator)
	}
	v := base64.StdEncoding.EncodeToString(entry.Value)
	return fmt.Sprintf("%s%s%s\n", k, Seperator, v), nil
}
