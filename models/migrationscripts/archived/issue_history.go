package archived

import (
	"time"
)

type IssueStatusHistory struct {
	NoPKModel
	IssueId        string    `gorm:"primaryKey;type:varchar(255)"`
	OriginalStatus string    `gorm:"primaryKey;type:varchar(255)"`
	StartDate      time.Time `gorm:"primaryKey"`
	EndDate        *time.Time
}

func (IssueStatusHistory) TableName() string {
	return "issue_status_history"
}

type IssueAssigneeHistory struct {
	NoPKModel
	IssueId   string    `gorm:"primaryKey;type:varchar(255)"`
	Assignee  string    `gorm:"primaryKey;type:varchar(255)"`
	StartDate time.Time `gorm:"primaryKey"`
	EndDate   *time.Time
}

func (IssueAssigneeHistory) TableName() string {
	return "issue_assignee_history"
}

type IssueSprintsHistory struct {
	NoPKModel
	IssueId   string    `gorm:"primaryKey;type:varchar(255)"`
	SprintId  string    `gorm:"primaryKey;type:varchar(255)"`
	StartDate time.Time `gorm:"primaryKey"`
	EndDate   *time.Time
}

func (IssueSprintsHistory) TableName() string {
	return "issue_sprints_history"
}
