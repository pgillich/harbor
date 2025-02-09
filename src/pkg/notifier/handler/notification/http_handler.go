package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// HTTPHandler preprocess http event data and start the hook processing
type HTTPHandler struct {
}

// Name ...
func (h *HTTPHandler) Name() string {
	return "HTTP"
}

// Handle handles http event
func (h *HTTPHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("HTTPHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification http event")
	}
	return h.process(ctx, event)
}

// IsStateful ...
func (h *HTTPHandler) IsStateful() bool {
	return false
}

func (h *HTTPHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.WebhookJobVendorType

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal from payload %v failed: %v", event.Payload, err)
	}

	j.Parameters = map[string]interface{}{
		"payload": string(payload),
		"address": event.Target.Address,
		// Users can define a auth header in http statement in notification(webhook) policy.
		// So it will be sent in header in http request.
		"auth_header":      event.Target.AuthHeader,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}
