package services

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	texttemplate "text/template"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type OpsgenieOptions struct {
	ApiUrl  string            `json:"apiUrl"`
	ApiKeys map[string]string `json:"apiKeys"`
}

type OpsgenieNotification struct {
	Alias       string            `json:"alias,omitempty"`
	Description string            `json:"description,omitempty"`
	VisibleTo   []alert.Responder `json:"visibleTo,omitempty"`
	Actions     []string          `json:"actions,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	Entity      string            `json:"entity,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	User        string            `json:"user,omitempty"`
	Note        string            `json:"note,omitempty"`
}

func (n *OpsgenieNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	alias, err := texttemplate.New(name).Funcs(f).Parse(n.Alias)
	if err != nil {
		return nil, err
	}
	desc, err := texttemplate.New(name).Funcs(f).Parse(n.Description)
	if err != nil {
		return nil, err
	}
	//visibleTo, err := texttemplate.New(name).Funcs(f).Parse(n.VisibleTo)
	//if err != nil {
	//	return nil, err
	//}
	//actions, err := texttemplate.New(name).Funcs(f).Parse(n.Actions)
	//if err != nil {
	//	return nil, err
	//}
	//tags, err := texttemplate.New(name).Funcs(f).Parse(n.Tags)
	//if err != nil {
	//	return nil, err
	//}
	//details, err := texttemplate.New(name).Funcs(f).Parse(n.Details)
	//if err != nil {
	//	return nil, err
	//}
	entity, err := texttemplate.New(name).Funcs(f).Parse(n.Entity)
	if err != nil {
		return nil, err
	}
	priority, err := texttemplate.New(name).Funcs(f).Parse(n.Priority)
	if err != nil {
		return nil, err
	}
	user, err := texttemplate.New(name).Funcs(f).Parse(n.User)
	if err != nil {
		return nil, err
	}
	note, err := texttemplate.New(name).Funcs(f).Parse(n.Note)
	if err != nil {
		return nil, err
	}
	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Opsgenie == nil {
			notification.Opsgenie = &OpsgenieNotification{}
		}
		var aliasData, descData, entityData, priorityData, userData, noteData bytes.Buffer

		if err := alias.Execute(&aliasData, vars); err != nil {
			return err
		}
		if err := desc.Execute(&descData, vars); err != nil {
			return err
		}
		//if err := visibleTo.Execute(&visibleToData, vars); err != nil {
		//	return err
		//}
		//if err := actions.Execute(&actionsData, vars); err != nil {
		//	return err
		//}
		//if err := tags.Execute(&tagsData, vars); err != nil {
		//	return err
		//}
		//if err := details.Execute(&detailsData, vars); err != nil {
		//	return err
		//}
		if err := entity.Execute(&entityData, vars); err != nil {
			return err
		}
		if err := priority.Execute(&priorityData, vars); err != nil {
			return err
		}
		if err := user.Execute(&userData, vars); err != nil {
			return err
		}
		if err := note.Execute(&noteData, vars); err != nil {
			return err
		}

		notification.Opsgenie.Alias = aliasData.String()
		notification.Opsgenie.Description = descData.String()
		//notification.Opsgenie.VisibleTo = visibleToData.String()
		notification.Opsgenie.VisibleTo = n.VisibleTo
		//notification.Opsgenie.Actions = actionsData.String()
		notification.Opsgenie.Actions = n.Actions
		//notification.Opsgenie.Tags = tagsData.String()
		notification.Opsgenie.Tags = n.Tags
		notification.Opsgenie.Details = n.Details
		//notification.Opsgenie.Details = detailsData.String()
		notification.Opsgenie.Entity = entityData.String()
		notification.Opsgenie.Priority = priorityData.String()
		notification.Opsgenie.User = userData.String()
		notification.Opsgenie.Note = noteData.String()
		return nil
	}, nil
}

type opsgenieService struct {
	opts OpsgenieOptions
}

func NewOpsgenieService(opts OpsgenieOptions) NotificationService {
	return &opsgenieService{opts: opts}
}

func (s *opsgenieService) Send(notification Notification, dest Destination) error {
	apiKey, ok := s.opts.ApiKeys[dest.Recipient]
	if !ok {
		return fmt.Errorf("no API key configured for recipient %s", dest.Recipient)
	}
	alertClient, _ := alert.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(s.opts.ApiUrl),
		HttpClient: &http.Client{
			Transport: httputil.NewLoggingRoundTripper(
				httputil.NewTransport(s.opts.ApiUrl, false), log.WithField("service", "opsgenie")),
		},
	})
	description := ""
	alias := ""
	entity := ""
	priority := ""
	user := ""
	note := ""
	actions := []string(nil)
	tags := []string(nil)
	details := map[string]string(nil)
	var visibleTo []alert.Responder

	if notification.Opsgenie != nil {
		description = notification.Opsgenie.Description
		alias = notification.Opsgenie.Alias
		entity = notification.Opsgenie.Entity
		priority = notification.Opsgenie.Priority
		user = notification.Opsgenie.User
		note = notification.Opsgenie.Note
		actions = notification.Opsgenie.Actions
		tags = notification.Opsgenie.Tags
		details = notification.Opsgenie.Details
		visibleTo = notification.Opsgenie.VisibleTo
	}

	_, err := alertClient.Create(context.TODO(), &alert.CreateAlertRequest{
		Message:     notification.Message,
		Alias:       alias,
		Description: description,
		Responders: []alert.Responder{
			{
				Type: "team",
				Id:   dest.Recipient,
			},
		},
		Actions:   actions,
		Tags:      tags,
		Details:   details,
		Entity:    entity,
		Source:    "Argo CD",
		Priority:  alert.Priority(priority),
		User:      user,
		Note:      note,
		VisibleTo: visibleTo,
	})
	return err
}
