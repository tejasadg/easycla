// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v32/github"

	"github.com/communitybridge/easycla/cla-backend-go/logging"
)

// GetInstallationRepositories returns list of repositories for github app installation
func GetInstallationRepositories(installationID int64) ([]*github.Repository, error) {
	client, err := NewGithubAppClient(installationID)
	if err != nil {
		return nil, errors.New("cannot create github client")
	}
	repos, _, err := client.Apps.ListRepos(context.TODO(), nil)
	if err != nil {
		logging.Error("error while getting installation repositories", err)
		err = fmt.Errorf("unable to get repositories for installation id : %d", installationID)
		return nil, err
	}
	return repos, nil
}
