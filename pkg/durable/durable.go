package durable

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/durableio/cli/pkg/cache"
	"github.com/durableio/cli/pkg/logging"
)

type WorkflowId string
type Workflow struct {
	sync.RWMutex
	WorkflowId WorkflowId
	StepIds    []StepId
}

type StepId string
type Step struct {
	sync.RWMutex
	WorkflowId  WorkflowId
	Id          StepId
	Name        string
	CallbackUrl string
	Body        struct {
		Method string
		Url    string
		Header http.Header
		Body   string
	}
	Done   bool
	Result struct {
		Header http.Header
		Body   string
	}
}

type durable struct {
	close       chan struct{}
	logger      logging.Logger
	stepCounter atomic.Int64
	queue       chan *Step
	steps       cache.Cache[*Step]
	workflows   cache.Cache[*Workflow]
}

func (d *durable) Enqueue(ctx context.Context, step *Step) error {
	if !d.workflows.Contains(string(step.WorkflowId)) {
		d.workflows.Set(string(step.WorkflowId), &Workflow{WorkflowId: step.WorkflowId, StepIds: []StepId{}})
	}

	wf, err := d.workflows.Get(string(step.WorkflowId))
	if err != nil {
		return err
	}
	wf.Lock()
	defer wf.Unlock()
	wf.StepIds = append(wf.StepIds, step.Id)
	d.steps.Set(string(step.Id), step)
	d.queue <- step
	d.stepCounter.Add(1)
	return nil
}
func (d *durable) GetWorkflow(workflowId WorkflowId) (*Workflow, error) {
	return d.workflows.Get(string(workflowId))
}
func (d *durable) GetStep(stepId StepId) (*Step, error) {
	return d.steps.Get(string(stepId))
}

func (d *durable) Run() error {
	for {
		select {
		case <-d.close:
			return nil

		case step := <-d.queue:
			logger := d.logger.With().Str("workflowId", string(step.WorkflowId)).Str("stepId", string(step.Id)).Logger()
			logger.Info().Msg("Handling step")
			buf := bytes.NewBuffer(nil)

			_, err := buf.WriteString(step.Body.Body)
			if err != nil {
				return err
			}

			req, err := http.NewRequest(step.Body.Method, step.Body.Url, buf)
			if err != nil {
				return err
			}
			for k, vs := range step.Body.Header {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			logger.Debug().Str("url", step.Body.Url).Msg("Calling")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			logger.Debug().Int("status", res.StatusCode).Msg("Response")

			step.Done = true
			step.Result.Header = res.Header.Clone()
			b, err := io.ReadAll(res.Body)
			closeErr := res.Body.Close()
			if closeErr != nil {
				return closeErr
			}
			if err != nil {
				return err
			}

			logger.Debug().Interface("step", step).Send()
			step.Result.Body = string(b)
			if step.CallbackUrl != "" {
				logger.Info().Msg("Preparing callback")
				wf, err := d.workflows.Get(string(step.WorkflowId))
				if err != nil {
					return err
				}
				body := make(map[string]any)
				wf.RLock()
				stepIds := wf.StepIds
				wf.RUnlock()
				for _, stepId := range stepIds {
					step, err := d.steps.Get(string(stepId))
					if err != nil {
						return err
					}
					step.RLock()

					body[step.Name] = step.Result.Body
					step.RUnlock()
				}
				logger.Info().Str("callback", step.CallbackUrl).Msg("Sending callback")
				buf, err := json.Marshal(body)
				if err != nil {
					return err
				}

				cbReq, err := http.NewRequest("POST", step.CallbackUrl, bytes.NewBuffer(buf))
				if err != nil {
					return err
				}
				cbReq.Header.Set("User-Agent", "durable.io")
				cbReq.Header.Set("Content-Type", "application/json")
				cbReq.Header.Set("Durable-Workflow-Id", string(wf.WorkflowId))
				cbReq.Header.Set("Durable-Callback", step.CallbackUrl)
				cbRes, err := http.DefaultClient.Do(cbReq)
				if err != nil {
					return err
				}
				cbReq.Body.Close()
				if cbRes.StatusCode >= 400 {
					logger.Warn().Str("callback", step.CallbackUrl).Str("status", cbRes.Status).Msg("Callback failed")
				}

			}

		}
	}
}

func (d *durable) Close() {
	d.close <- struct{}{}
}

type Durable interface {
	Enqueue(ctx context.Context, step *Step) error
	Run() error
	GetWorkflow(workflowId WorkflowId) (*Workflow, error)
	GetStep(stepId StepId) (*Step, error)
}

type Config struct {
	Logger logging.Logger
}

func New(cfg Config) Durable {
	return &durable{
		close:       make(chan struct{}),
		logger:      cfg.Logger,
		stepCounter: atomic.Int64{},
		queue:       make(chan *Step, 128),
		steps:       cache.NewInMemoryCache[*Step](time.Hour),
		workflows:   cache.NewInMemoryCache[*Workflow](time.Hour),
	}
}
