package server

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type step struct {
	WorkflowId string `json:"workflowId"`
	Id         string `json:"id"`
	Name       string `json:"name"`

	Done   bool `json:"done"`
	Result struct {
		Header http.Header `json:"header"`
		Body   string      `json:"body"`
	} `json:"result,omitempty"`
}

type PollResponse struct {
	Steps []step `json:"steps"`
}

func (s *Server) workflow(c *fiber.Ctx) error {

	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	wf, err := s.durable.GetWorkflowFromToken(token)
	if err != nil {
		return fiber.NewError(fiber.ErrInternalServerError.Code, err.Error())
	}

	steps := []step{}
	for _, stepId := range wf.StepIds {
		s, err := s.durable.GetStep(stepId)
		if err != nil {
			return fiber.NewError(fiber.ErrInternalServerError.Code, err.Error())
		}

		ss := step{
			WorkflowId: string(s.WorkflowId),
			Id:         string(s.StepId),
			Name:       s.Name,
			Done:       s.Done,
		}
		ss.Result.Header = s.Result.Header.Clone()
		ss.Result.Body = s.Result.Body
		steps = append(steps, ss)
	}

	return c.JSON(PollResponse{Steps: steps})

}
