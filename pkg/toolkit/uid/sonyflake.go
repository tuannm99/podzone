package uid

import (
	"github.com/sony/sonyflake"
)

func New() (uint64, error) {
	return sonyflake.NewSonyflake(sonyflake.Settings{}).NextID()
}
