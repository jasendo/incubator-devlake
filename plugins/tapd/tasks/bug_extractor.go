package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/merico-dev/lake/models/domainlayer/ticket"
	"github.com/merico-dev/lake/plugins/core"
	"github.com/merico-dev/lake/plugins/helper"
	"github.com/merico-dev/lake/plugins/tapd/models"
	"strings"
)

var _ core.SubTaskEntryPoint = ExtractBugs

var ExtractBugMeta = core.SubTaskMeta{
	Name:             "extractBugs",
	EntryPoint:       ExtractBugs,
	EnabledByDefault: true,
	Description:      "Extract raw workspace data into tool layer table _tool_tapd_iterations",
}

type TapdBugRes struct {
	Bug models.TapdBug
}

func ExtractBugs(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*TapdTaskData)
	db := taskCtx.GetDb()
	statusList := make([]*models.TapdBugStatus, 0)

	err := db.Model(&models.TapdBugStatus{}).
		Find(&statusList, "connection_id = ? and workspace_id = ?", data.Options.ConnectionId, data.Options.WorkspaceID).
		Error
	if err != nil {
		return err
	}

	statusMap := make(map[string]string, len(statusList))
	for _, v := range statusList {
		statusMap[v.EnglishName] = v.ChineseName
	}

	getStdStatus := func(statusKey string) string {
		if statusKey == "已关闭" || statusKey == "不处理" {
			return ticket.DONE
		} else if statusKey == "新建" {
			return ticket.TODO
		} else {
			return ticket.IN_PROGRESS
		}
	}
	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: TapdApiParams{
				ConnectionId: data.Connection.ID,
				//CompanyId: data.Options.CompanyId,
				WorkspaceID: data.Options.WorkspaceID,
			},
			Table: RAW_BUG_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			var bugBody TapdBugRes
			err := json.Unmarshal(row.Data, &bugBody)
			if err != nil {
				return nil, err
			}
			toolL := bugBody.Bug

			toolL.Status = statusMap[toolL.Status]
			toolL.ConnectionId = data.Connection.ID
			toolL.Type = "BUG"
			toolL.StdType = "BUG"
			toolL.StdStatus = getStdStatus(toolL.Status)
			toolL.Url = fmt.Sprintf("https://www.tapd.cn/%d/prong/stories/view/%d", toolL.WorkspaceID, toolL.ID)
			if strings.Contains(toolL.CurrentOwner, ";") {
				toolL.CurrentOwner = strings.Split(toolL.CurrentOwner, ";")[0]
			}
			workSpaceBug := &models.TapdWorkSpaceBug{
				ConnectionId: data.Connection.ID,
				WorkspaceID:  toolL.WorkspaceID,
				BugId:        toolL.ID,
			}
			results := make([]interface{}, 0, 3)
			results = append(results, &toolL, workSpaceBug)
			if toolL.IterationID != 0 {
				iterationBug := &models.TapdIterationBug{
					ConnectionId:   data.Connection.ID,
					IterationId:    toolL.IterationID,
					WorkspaceID:    toolL.WorkspaceID,
					BugId:          toolL.ID,
					ResolutionDate: toolL.Resolved,
					BugCreatedDate: toolL.Created,
				}
				results = append(results, iterationBug)
			}
			if toolL.Label != "" {
				labelList := strings.Split(toolL.Label, "|")
				for _, v := range labelList {
					toolLIssueLabel := &models.TapdBugLabel{
						BugId:     toolL.ID,
						LabelName: v,
					}
					results = append(results, toolLIssueLabel)
				}
			}
			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
