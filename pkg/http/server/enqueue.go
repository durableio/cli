package server

import (
	"fmt"
	"net/http"

	"github.com/durableio/cli/pkg/durable"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type EnqueueRequest struct {
	Method string            `json:"method" validate:"required"`
	Url    string            `json:"url" validate:"required,url"`
	Header map[string]string `json:"header,omitempty"`
	Body   string            `json:"body,omitempty"`

	WorkflowId durable.WorkflowId `json:"workflowId,omitempty"`

	StepName string `json:"stepName" validate:"required"`

	CallbackUrl string `json:"callbackUrl,omitempty"`
}

func (r EnqueueRequest) validate() error {
	err := validator.New().Struct(r)
	if err != nil {
		return err
	}
	switch r.Method {
	case "GET":
	case "POST":
		return nil
	default:
		return fmt.Errorf("Unhandled method: %s", r.Method)
	}

	return nil

}

type EnqueueResponse struct {
	ReadWorkflowToken string             `json:"readWorkflowToken"`
	WorkflowId        durable.WorkflowId `json:"workflowId"`
	StepId            durable.StepId     `json:"stepId"`
}

func (s *Server) enqueue(c *fiber.Ctx) error {

	req := &EnqueueRequest{}
	err := c.BodyParser(req)
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, err.Error())
	}
	err = req.validate()
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, err.Error())
	}
	s.logger.Debug().Interface("req", req).Send()
	step := durable.EnqueueRequest{}
	step.WorkflowId = req.WorkflowId

	step.Body.Method = req.Method
	step.StepName = req.StepName
	step.CallbackUrl = req.CallbackUrl
	step.Body.Url = req.Url
	step.Body.Header = http.Header{}
	for k, v := range req.Header {
		step.Body.Header.Add(k, v)
	}
	step.Body.Body = req.Body
	s.logger.Info().Str("step", step.StepName).Msg("Received")
	res, err := s.durable.Enqueue(c.UserContext(), step)
	if err != nil {
		return fiber.NewError(fiber.ErrInternalServerError.Code, err.Error())
	}

	return c.JSON(EnqueueResponse{
		ReadWorkflowToken: res.ReadWorkflowToken,
		WorkflowId:        res.WorkflowId,
		StepId:            res.StepId,
	})

}
