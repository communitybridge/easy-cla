package gitlab

import (
	"fmt"
	"testing"

	gitlab_api "github.com/communitybridge/easycla/cla-backend-go/gitlab_api"
	"github.com/communitybridge/easycla/cla-backend-go/gitlab_api/mocks"
	"github.com/communitybridge/easycla/cla-backend-go/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	goGitLab "github.com/xanzy/go-gitlab"
)

func TestFetchMrParticipants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGitLabClient := mocks.NewMockGitLabClient(ctrl)

	projectID := 123
	mergeID := 456

	commits := []*goGitLab.Commit{
		{
			ID:          "commit1",
			AuthorEmail: "author1@example.com",
			AuthorName:  "Author 1",
			ShortID:     "shortID1",
		},
		{
			ID:          "commit2",
			AuthorEmail: "author2@example.com",
			AuthorName:  "Author 2",
			ShortID:     "shortID2",
		},
	}

	users := []*goGitLab.User{
		{
			ID:       1,
			Username: "author_username1",
			Email:    "author1@example.com",
			Name:     "Author 1",
		},
		{
			ID:       2,
			Username: "author_username2",
			Email:    "author2@example.com",
			Name:     "Author 2",
		},
	}

	testCases := []struct {
		name            string
		mockSetup       func()
		expectedResults []*gitlab_api.UserCommitSummary
		expectedError   error
	}{
		{
			name: "Successful fetch with participants",
			mockSetup: func() {
				mockGitLabClient.EXPECT().GetMergeRequestCommits(projectID, mergeID, &goGitLab.GetMergeRequestCommitsOptions{}).Return(commits, nil)
				for index, commit := range commits {
					mockGitLabClient.EXPECT().ListUsers(&goGitLab.ListUsersOptions{Active: utils.Bool(true), Blocked: utils.Bool(false), Search: &commit.AuthorEmail}).Return([]*goGitLab.User{users[index]}, nil)
				}
			},
			expectedResults: []*gitlab_api.UserCommitSummary{
				{
					CommitSha:      "shortID1",
					AuthorName:     "Author 1",
					AuthorEmail:    "author1@example.com",
					Authorized:     false,
					Affiliated:     false,
					AuthorID:       1,
					AuthorUsername: "author_username1",
				},
				{
					CommitSha:      "shortID2",
					AuthorName:     "Author 2",
					AuthorEmail:    "author2@example.com",
					Authorized:     false,
					Affiliated:     false,
					AuthorID:       2,
					AuthorUsername: "author_username2",
				},
			},
			expectedError: nil,
		},
		{
			name: "No commits found",
			mockSetup: func() {
				mockGitLabClient.EXPECT().GetMergeRequestCommits(projectID, mergeID, &goGitLab.GetMergeRequestCommitsOptions{}).Return([]*goGitLab.Commit{}, nil)
			},
			expectedResults: []*gitlab_api.UserCommitSummary{},
			expectedError:   nil,
		},
		{
			name: "Error fetching commits",
			mockSetup: func() {
				mergeRequestError := fmt.Errorf("error fetching merge request commits")
				mockGitLabClient.EXPECT().GetMergeRequestCommits(projectID, mergeID, &goGitLab.GetMergeRequestCommitsOptions{}).Return(nil, mergeRequestError)
			},
			expectedResults: nil,
			expectedError:   fmt.Errorf("fetching gitlab participants for project : %d and merge id : %d, failed : %v", projectID, mergeID, fmt.Errorf("error fetching merge request commits")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			results, err := gitlab_api.FetchMrParticipants(mockGitLabClient, projectID, mergeID)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResults, results)
		})
	}
}
