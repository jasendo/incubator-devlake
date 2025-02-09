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

var _ core.SubTaskEntryPoint = ExtractTasks

var ExtractTaskMeta = core.SubTaskMeta{
	Name:             "extractTasks",
	EntryPoint:       ExtractTasks,
	EnabledByDefault: true,
	Description:      "Extract raw workspace data into tool layer table _tool_tapd_iterations",
}

type TapdTaskRes struct {
	Task models.TapdTask
}

func ExtractTasks(taskCtx core.SubTaskContext) error {
	data := taskCtx.GetData().(*TapdTaskData)
	getStdStatus := func(statusKey string) string {
		if statusKey == "done" {
			return ticket.DONE
		} else if statusKey == "progressing" {
			return ticket.IN_PROGRESS
		} else {
			return ticket.TODO
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
			Table: RAW_TASK_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			var taskBody TapdTaskRes
			err := json.Unmarshal(row.Data, &taskBody)
			if err != nil {
				return nil, err
			}
			toolL := taskBody.Task

			toolL.ConnectionId = data.Connection.ID
			toolL.Type = "TASK"
			toolL.StdType = "TASK"
			toolL.StdStatus = getStdStatus(toolL.Status)
			if strings.Contains(toolL.Owner, ";") {
				toolL.Owner = strings.Split(toolL.Owner, ";")[0]
			}
			toolL.Url = fmt.Sprintf("https://www.tapd.cn/%d/prong/stories/view/%d", toolL.WorkspaceID, toolL.ID)

			workSpaceTask := &models.TapdWorkSpaceTask{
				ConnectionId: data.Connection.ID,
				WorkspaceID:  toolL.WorkspaceID,
				TaskId:       toolL.ID,
			}
			results := make([]interface{}, 0, 3)
			results = append(results, &toolL, workSpaceTask)
			if toolL.IterationID != 0 {
				iterationTask := &models.TapdIterationTask{
					ConnectionId:    data.Connection.ID,
					IterationId:     toolL.IterationID,
					TaskId:          toolL.ID,
					WorkspaceID:     toolL.WorkspaceID,
					ResolutionDate:  toolL.Completed,
					TaskCreatedDate: toolL.Created,
				}
				results = append(results, iterationTask)
			}
			if toolL.Label != "" {
				labelList := strings.Split(toolL.Label, "|")
				for _, v := range labelList {
					toolLIssueLabel := &models.TapdTaskLabel{
						TaskId:    toolL.ID,
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
