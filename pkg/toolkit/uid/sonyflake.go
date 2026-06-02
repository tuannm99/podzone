package uid

import (
	"fmt"

	"github.com/sony/sonyflake"
)

func New() (uint64, error) {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return 1, nil
		},
	})
	if sf == nil {
		return 0, fmt.Errorf("sonyflake: failed to initialize")
	}
	return sf.NextID()
}
