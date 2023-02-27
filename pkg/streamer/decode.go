package streamer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"keepair/pkg/common"
)

func DecodeMessage(line string) (common.Entry, error) {
	parts := strings.SplitN(line, Seperator, 2)
	if len(parts) != 2 {
		return common.Entry{}, errors.New("line has invalid number of segments")
	}
	k := parts[0]
	v, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return common.Entry{}, fmt.Errorf("failed to decode message: %w", err)
	}
	return common.Entry{
		Key:   k,
		Value: v,
	}, nil
}
