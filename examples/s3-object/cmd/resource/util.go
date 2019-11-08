package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws/session"
)

type handlerFunc func(*Model, *Model, *session.Session) (handler.ProgressEvent, error)

func makeError(err error) handler.ProgressEvent {
	return handler.NewFailedEvent(cfnerr.New(
		cfnerr.GeneralServiceException,
		"Unable to complete request: "+err.Error(),
		err,
	))
}

func wrap(ctx context.Context, req handler.Request, f handlerFunc) (response handler.ProgressEvent) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = errors.New(fmt.Sprint(r))
			}

			response = makeError(err)
		}
	}()

	// Populate the previous model
	var prevModel *Model
	if err := req.UnmarshalPrevious(prevModel); err != nil {
		return makeError(err)
	}

	// Populate the current model
	currentModel := &Model{}
	if err := req.Unmarshal(currentModel); err != nil {
		return makeError(err)
	}

	// Retrieve the session
	session, err := cfn.GetContextSession(ctx)
	if err != nil {
		return makeError(err)
	}

	response, err = f(prevModel, currentModel, session)
	if err != nil {
		return makeError(err)
	}

	return response
}
