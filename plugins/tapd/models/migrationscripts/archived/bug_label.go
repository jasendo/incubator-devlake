package archived

import (
	"github.com/merico-dev/lake/models/common"
)

type TapdBugLabel struct {
	BugId     uint64 `gorm:"primaryKey;autoIncrement:false"`
	LabelName string `gorm:"primaryKey;type:varchar(255)"`
	common.NoPKModel
}

func (TapdBugLabel) TableName() string {
	return "_tool_tapd_bug_labels"
}
