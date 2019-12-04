package resource

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/encoding"
	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func makeGitHubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(context.Background(), ts)

	return github.NewClient(tc)
}

func parseURL(url string) (string, string) {
	parts := strings.Split(url, "/")

	return parts[len(parts)-2], parts[len(parts)-1]
}

// Create a repo
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	client := makeGitHubClient(*currentModel.OauthToken.Value())

	log.Printf("Attempting to create repository: %s/%s", *currentModel.Owner.Value(), *currentModel.Name.Value())

	repo, resp, err := client.Repositories.Create(context.Background(), "", &github.Repository{
		Name:        currentModel.Name.Value(),
		Homepage:    currentModel.Homepage.Value(),
		Description: currentModel.Description.Value(),
		Owner: &github.User{
			Name: currentModel.Owner.Value(),
		},
	})

	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 201 {
		log.Printf("Got a non-201 error code: %v", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	currentModel.URL = encoding.NewString(repo.GetURL())

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create Complete",
		ResourceModel:   currentModel,
	}, nil
}

// Read a repo status
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL.Value())

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner.Value(), *currentModel.Name.Value())
	client := makeGitHubClient(*currentModel.OauthToken.Value())
	repo, resp, err := client.Repositories.Get(context.Background(), owner, repoName)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Unable to find repository: %s", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	currentModel.Name = encoding.NewString(*repo.Name)
	currentModel.Owner = encoding.NewString(*repo.Owner.Name)
	currentModel.Description = encoding.NewString(*repo.Description)
	currentModel.Homepage = encoding.NewString(*repo.Homepage)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Read Complete",
		ResourceModel:   currentModel,
	}, nil
}

// Update a repo
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL.Value())

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner.Value(), *currentModel.Name.Value())
	client := makeGitHubClient(*currentModel.OauthToken.Value())

	_, resp, err := client.Repositories.Edit(context.Background(), owner, repoName, &github.Repository{
		Homepage:    currentModel.Homepage.Value(),
		Description: currentModel.Description.Value(),
	})
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Unable to find repository: %s", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update Complete",
		ResourceModel:   currentModel,
	}, nil
}

// Delete a repo
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL.Value())

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner.Value(), *currentModel.Name.Value())
	client := makeGitHubClient(*currentModel.OauthToken.Value())

	resp, err := client.Repositories.Delete(context.Background(), owner, repoName)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Unable to find repository: %s", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Delete Complete",
		ResourceModel:   currentModel,
	}, nil
}

// List ...
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// no-op

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "List Complete",
		ResourceModel:   currentModel,
	}, nil
}
