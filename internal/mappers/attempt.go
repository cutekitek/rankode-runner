package mappers

import (
	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/repository/models"
)

func RunResultToAttemptResult(req *models.AttemptRequest, result *dto.RunResult) *models.AttemptResponse {
	resp := &models.AttemptResponse{
		Id:          req.Id,
		Status:      result.Status,
		MemoryUsage: int64(result.MemoryUsage),
		Tests:       make([]models.TestStatus, 0, len(result.Output)),
	}
	for i, out := range result.Output {
		status := models.TestStatus{CaseId: req.TestCases[i].Id, Status: out.Status, Output: out.Output}
		resp.Tests = append(resp.Tests, status)
	}
	return resp
}
