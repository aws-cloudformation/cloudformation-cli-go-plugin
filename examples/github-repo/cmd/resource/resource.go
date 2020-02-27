package resource

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go/aws"
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

// Create handles the Create event from the Cloudformation service.
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	client := makeGitHubClient(*currentModel.OauthToken)

	log.Printf("Attempting to create repository: %s/%s", *currentModel.Owner, *currentModel.Name)

	repo, resp, err := client.Repositories.Create(context.Background(), "", &github.Repository{
		Name:        currentModel.Name,
		Homepage:    currentModel.Homepage,
		Description: currentModel.Description,
		Owner: &github.User{
			Name: currentModel.Owner,
		},
	})

	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 201 {
		log.Printf("Got a non-201 error code: %v", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	currentModel.URL = aws.String(repo.GetURL())

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create Complete",
		ResourceModel:   currentModel,
	}, nil
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL)

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner, *currentModel.Name)
	client := makeGitHubClient(*currentModel.OauthToken)
	repo, resp, err := client.Repositories.Get(context.Background(), owner, repoName)
	if err != nil {
		return handler.ProgressEvent{}, err
	}

	if resp.StatusCode != 200 {
		log.Printf("Unable to find repository: %s", resp.Status)
		return handler.ProgressEvent{}, fmt.Errorf("Status Code: %d, Status: %v", resp.StatusCode, resp.Status)
	}

	currentModel.Name = aws.String(*repo.Name)
	currentModel.Owner = aws.String(*repo.Owner.Name)
	currentModel.Description = aws.String(*repo.Description)
	currentModel.Homepage = aws.String(*repo.Homepage)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Read Complete",
		ResourceModel:   currentModel,
	}, nil
}

// Update handles the Update event from the Cloudformation service.
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL)

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner, *currentModel.Name)
	client := makeGitHubClient(*currentModel.OauthToken)

	_, resp, err := client.Repositories.Edit(context.Background(), owner, repoName, &github.Repository{
		Homepage:    currentModel.Homepage,
		Description: currentModel.Description,
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

// Delete handles the Delete event from the Cloudformation service.
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	owner, repoName := parseURL(*currentModel.URL)

	log.Printf("Looking for repository: %s/%s", *currentModel.Owner, *currentModel.Name)
	client := makeGitHubClient(*currentModel.OauthToken)

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

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "List Complete",
		ResourceModel:   currentModel,
	}, nil
}
