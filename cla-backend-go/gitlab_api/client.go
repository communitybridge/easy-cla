// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package gitlab

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"

	goGitLab "github.com/xanzy/go-gitlab"
)

type GitLabClient interface {
	GetMergeRequestCommits(projectID int, mergeID int, opts *goGitLab.GetMergeRequestCommitsOptions) ([]*goGitLab.Commit, error)
	ListUsers(opts *goGitLab.ListUsersOptions) ([]*goGitLab.User, error)
	SetCommitStatus(projectID int, commitSHA string, opts *goGitLab.SetCommitStatusOptions) error
	GetProject(gitLabProjectID int, opts *goGitLab.GetProjectOptions) (*goGitLab.Project, error)
	// EditProject(projectID int,)
	ListGroupProjects(groupID int, opts *goGitLab.ListGroupProjectsOptions) ([]*goGitLab.Project, *goGitLab.Response, error)
	ListGroupMembers(gid interface{}, opt *goGitLab.ListGroupMembersOptions) ([]*goGitLab.GroupMember, *goGitLab.Response, error)
	CurrentUser() (*goGitLab.User, *goGitLab.Response, error)
	ListUserProjects(user interface{}, opt *goGitLab.ListProjectsOptions) ([]*goGitLab.Project, *goGitLab.Response, error)
	DeleteProjectHook(projectID, webhookID int) (*goGitLab.Response, error)
	ListProjectHooks(projectID int, opts *goGitLab.ListProjectHooksOptions) ([]*goGitLab.ProjectHook, *goGitLab.Response, error)
	AddProjectHook(projectID int, opts *goGitLab.AddProjectHookOptions) (*goGitLab.ProjectHook, *goGitLab.Response, error)
	EditProjectHook(projectID, existingID int, opts *goGitLab.EditProjectHookOptions) (*goGitLab.ProjectHook, *goGitLab.Response, error)
}

// Client is the gitlab client
type GitLabClientWrapper struct {
	gitlabClient *goGitLab.Client
}

// OauthSuccessResponse is success response from Gitlab
type OauthSuccessResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int    `json:"created_at"`
}

// NewGitlabOauthClient creates a new gitlab client from the given oauth info, authInfo is encrypted
func NewGitlabOauthClient(authInfo string, gitLabApp *App) (GitLabClient, error) {
	if authInfo == "" {
		return nil, errors.New("unable to decrypt auth info - authentication info input is nil")
	}
	if gitLabApp == nil || gitLabApp.gitLabAppID == "" || gitLabApp.gitLabAppPrivateKey == "" || gitLabApp.gitLabAppSecret == "" {
		return nil, errors.New("unable to decrypt auth info - GitLab app structure is nil or empty")
	}

	oauthResp, err := DecryptAuthInfo(authInfo, gitLabApp)
	if err != nil {
		return nil, err
	}

	if oauthResp == nil {
		return nil, errors.New("unable to decrypt auth info - value is nil")
	}

	log.Infof("creating oauth client with access token : %s", oauthResp.AccessToken)
	client, err := goGitLab.NewOAuthClient(oauthResp.AccessToken)
	if err != nil {
		return nil, err
	}

	return &GitLabClientWrapper{
		gitlabClient: client,
	}, nil
}

// NewGitlabOauthClientFromAccessToken creates a new gitlab client from the given access token
func NewGitlabOauthClientFromAccessToken(accessToken string) (*goGitLab.Client, error) {
	return goGitLab.NewOAuthClient(accessToken)
}

// EncryptAuthInfo encrypts the oauth response into a string
func EncryptAuthInfo(oauthResp *OauthSuccessResponse, gitLabApp *App) (string, error) {
	keyDecoded, err := base64.StdEncoding.DecodeString(gitLabApp.GetAppPrivateKey())
	if err != nil {
		return "", fmt.Errorf("problem decoding GitLab private glClientKey, error: %v", err)
	}

	b, err := json.Marshal(oauthResp)
	if err != nil {
		return "", fmt.Errorf("problem marshalling oauth resp json, error: %v", err)
	}
	authInfo := string(b)
	//log.Infof("auth info before encrypting : %s", authInfo)

	encrypted, err := encrypt(keyDecoded, []byte(authInfo))
	if err != nil {
		return "", fmt.Errorf("encrypt failed : %v", err)
	}

	return hex.EncodeToString(encrypted), nil
}

// DecryptAuthInfo decrypts the auth info into OauthSuccessResponse data structure
func DecryptAuthInfo(authInfoEncoded string, gitLabApp *App) (*OauthSuccessResponse, error) {
	ciphertext, err := hex.DecodeString(authInfoEncoded)
	if err != nil {
		return nil, fmt.Errorf("decode auth info %s : %v", authInfoEncoded, err)
	}

	//log.Infof("auth info decoded : %s", ciphertext)

	keyDecoded, err := base64.StdEncoding.DecodeString(gitLabApp.GetAppPrivateKey())
	if err != nil {
		return nil, fmt.Errorf("decode glClientKey : %v", err)
	}

	//log.Debugf("before decrypt : keyDecoded : %s, cipherText : %s", keyDecoded, ciphertext)
	decrypted, err := decrypt(keyDecoded, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed : %v", err)
	}
	//log.Debugf("after decrypt : keyDecoded : %s, decrypted : %s", keyDecoded, decrypted)

	var oauthResp OauthSuccessResponse
	if err := json.Unmarshal(decrypted, &oauthResp); err != nil {
		return nil, fmt.Errorf("unmarshall auth info : %v", err)
	}

	return &oauthResp, nil
}

func encrypt(key, message []byte) ([]byte, error) {
	// Initialize block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create the byte slice that will hold encrypted message
	cipherText := make([]byte, aes.BlockSize+len(message))

	// Generate the Initialization Vector (IV) nonce
	// which is stored at the beginning of the byte slice
	// The IV is the same length as the AES blocksize
	iv := cipherText[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	// Choose the block cipher mode of operation
	// Using the cipher feedback (CFB) mode here.
	// CBCEncrypter also available.
	cfb := cipher.NewCFBEncrypter(block, iv)
	// Generate the encrypted message and store it
	// in the remaining bytes after the IV nonce
	cfb.XORKeyStream(cipherText[aes.BlockSize:], message)

	return cipherText, nil
}

// AES decryption
func decrypt(key, cipherText []byte) ([]byte, error) {
	// Initialize block cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Separate the IV nonce from the encrypted message bytes
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Decrypt the message using the CFB block mode
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}

func (c *GitLabClientWrapper) ListGroupMembers(gid interface{}, opt *goGitLab.ListGroupMembersOptions) ([]*goGitLab.GroupMember, *goGitLab.Response, error) {
	return c.gitlabClient.Groups.ListGroupMembers(gid, opt)
}

func (c *GitLabClientWrapper) GetProject(gitLabProjectID int, opts *goGitLab.GetProjectOptions) (*goGitLab.Project, error) {
	project, _, err := c.gitlabClient.Projects.GetProject(gitLabProjectID, opts)
	if err != nil {
		return nil, err
	}
	return project, err
}

func (c *GitLabClientWrapper) CurrentUser() (*goGitLab.User, *goGitLab.Response, error) {
	return c.gitlabClient.Users.CurrentUser()
}

// GetMergeRequestCommits returns the commits for the given merge request
func (c *GitLabClientWrapper) GetMergeRequestCommits(projectID int, mergeID int, opts *goGitLab.GetMergeRequestCommitsOptions) ([]*goGitLab.Commit, error) {
	commits, _, err := c.gitlabClient.MergeRequests.GetMergeRequestCommits(projectID, mergeID, opts)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

// ListUsers returns the list of users
func (c *GitLabClientWrapper) ListUsers(opts *goGitLab.ListUsersOptions) ([]*goGitLab.User, error) {
	users, _, err := c.gitlabClient.Users.ListUsers(opts)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// SetCommitStatus sets the commit status
func (c *GitLabClientWrapper) SetCommitStatus(projectID int, commitSHA string, opts *goGitLab.SetCommitStatusOptions) error {
	_, _, err := c.gitlabClient.Commits.SetCommitStatus(projectID, commitSHA, opts)
	if err != nil {
		return err
	}
	return nil
}

func (c *GitLabClientWrapper) ListGroupProjects(groupID int, opts *goGitLab.ListGroupProjectsOptions) ([]*goGitLab.Project, *goGitLab.Response, error) {
	return c.gitlabClient.Groups.ListGroupProjects(groupID, opts)
}

func (c *GitLabClientWrapper) ListUserProjects(user interface{}, opt *goGitLab.ListProjectsOptions) ([]*goGitLab.Project, *goGitLab.Response, error) {
	return c.gitlabClient.Projects.ListUserProjects(user, opt)
}

func (c *GitLabClientWrapper) DeleteProjectHook(projectID, webhookID int) (*goGitLab.Response, error) {
	return c.gitlabClient.Projects.DeleteProjectHook(projectID, webhookID)
}

func (c *GitLabClientWrapper) ListProjectHooks(projectID int, opts *goGitLab.ListProjectHooksOptions) ([]*goGitLab.ProjectHook, *goGitLab.Response, error) {
	return c.gitlabClient.Projects.ListProjectHooks(projectID, opts)
}

func (c *GitLabClientWrapper) AddProjectHook(projectID int, opts *goGitLab.AddProjectHookOptions) (*goGitLab.ProjectHook, *goGitLab.Response, error) {
	return c.gitlabClient.Projects.AddProjectHook(projectID, opts)
}

func (c *GitLabClientWrapper) EditProjectHook(projectID, existingID int, opts *goGitLab.EditProjectHookOptions) (*goGitLab.ProjectHook, *goGitLab.Response, error) {
	return c.gitlabClient.Projects.EditProjectHook(projectID, existingID, opts)
}
