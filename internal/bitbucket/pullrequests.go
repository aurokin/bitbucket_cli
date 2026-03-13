package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ListPullRequestsOptions struct {
	State string
	Limit int
	Query string
	Sort  string
}

type CreatePullRequestOptions struct {
	Title             string
	Description       string
	SourceBranch      string
	DestinationBranch string
	CloseSourceBranch bool
	Draft             bool
	ReuseExisting     bool
}

type MergePullRequestOptions struct {
	Message           string
	CloseSourceBranch bool
	MergeStrategy     string
	PollInterval      time.Duration
	PollTimeout       time.Duration
}

type ListPullRequestTasksOptions struct {
	State string
	Limit int
	Query string
	Sort  string
}

type CreatePullRequestTaskOptions struct {
	Body      string
	CommentID int
	Pending   bool
}

type UpdatePullRequestTaskOptions struct {
	Body  string
	State string
}

type PullRequest struct {
	ID                int                      `json:"id"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	State             string                   `json:"state"`
	Author            PullRequestActor         `json:"author"`
	Reviewers         []PullRequestActor       `json:"reviewers,omitempty"`
	Participants      []PullRequestParticipant `json:"participants,omitempty"`
	Source            PullRequestRef           `json:"source"`
	Destination       PullRequestRef           `json:"destination"`
	CloseSourceBranch bool                     `json:"close_source_branch,omitempty"`
	Queued            bool                     `json:"queued,omitempty"`
	MergeCommit       PullRequestCommit        `json:"merge_commit,omitempty"`
	Draft             bool                     `json:"draft,omitempty"`
	CommentCount      int                      `json:"comment_count,omitempty"`
	TaskCount         int                      `json:"task_count,omitempty"`
	UpdatedOn         string                   `json:"updated_on,omitempty"`
	CreatedOn         string                   `json:"created_on,omitempty"`
	Links             PullRequestLinks         `json:"links,omitempty"`
}

type PullRequestActor struct {
	DisplayName string `json:"display_name,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
}

type PullRequestParticipant struct {
	User     PullRequestActor `json:"user"`
	Role     string           `json:"role,omitempty"`
	Approved bool             `json:"approved,omitempty"`
	State    string           `json:"state,omitempty"`
}

type PullRequestRef struct {
	Branch PullRequestBranch `json:"branch"`
	Commit PullRequestCommit `json:"commit"`
	Repo   PullRequestRepo   `json:"repository"`
}

type PullRequestBranch struct {
	Name                 string   `json:"name"`
	MergeStrategies      []string `json:"merge_strategies,omitempty"`
	DefaultMergeStrategy string   `json:"default_merge_strategy,omitempty"`
}

type PullRequestCommit struct {
	Hash string `json:"hash,omitempty"`
}

type PullRequestRepo struct {
	Name     string `json:"name,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

type PullRequestLinks struct {
	HTML Link `json:"html"`
}

type Link struct {
	Href string `json:"href"`
}

type pullRequestListResponse struct {
	Values []PullRequest `json:"values"`
	Next   string        `json:"next,omitempty"`
}

type PullRequestDiffStat struct {
	Status       string              `json:"status,omitempty"`
	Old          *PullRequestDiffRef `json:"old,omitempty"`
	New          *PullRequestDiffRef `json:"new,omitempty"`
	LinesAdded   int                 `json:"lines_added,omitempty"`
	LinesRemoved int                 `json:"lines_removed,omitempty"`
}

type PullRequestDiffRef struct {
	Path        string `json:"path,omitempty"`
	EscapedPath string `json:"escaped_path,omitempty"`
	Type        string `json:"type,omitempty"`
}

type pullRequestDiffStatResponse struct {
	Values []PullRequestDiffStat `json:"values"`
	Next   string                `json:"next,omitempty"`
}

type PullRequestComment struct {
	ID         int                        `json:"id"`
	Content    PullRequestCommentContent  `json:"content"`
	User       PullRequestActor           `json:"user"`
	Deleted    bool                       `json:"deleted,omitempty"`
	Parent     *PullRequestComment        `json:"parent,omitempty"`
	Inline     *PullRequestCommentInline  `json:"inline,omitempty"`
	Resolution *PullRequestCommentResolve `json:"resolution,omitempty"`
	Pending    bool                       `json:"pending,omitempty"`
	CreatedOn  string                     `json:"created_on,omitempty"`
	UpdatedOn  string                     `json:"updated_on,omitempty"`
	Links      PullRequestCommentLinks    `json:"links,omitempty"`
}

type PullRequestCommentContent struct {
	Raw string `json:"raw,omitempty"`
}

type PullRequestCommentInline struct {
	From      int    `json:"from,omitempty"`
	To        int    `json:"to,omitempty"`
	StartFrom int    `json:"start_from,omitempty"`
	StartTo   int    `json:"start_to,omitempty"`
	Path      string `json:"path,omitempty"`
}

type PullRequestCommentLinks struct {
	HTML Link `json:"html"`
}

type PullRequestCommentResolve struct {
	Type      string           `json:"type,omitempty"`
	User      PullRequestActor `json:"user,omitempty"`
	CreatedOn string           `json:"created_on,omitempty"`
}

type PullRequestTask struct {
	ID         int                    `json:"id"`
	State      string                 `json:"state,omitempty"`
	Content    PullRequestTaskContent `json:"content"`
	Creator    PullRequestActor       `json:"creator"`
	Pending    bool                   `json:"pending,omitempty"`
	ResolvedOn string                 `json:"resolved_on,omitempty"`
	ResolvedBy PullRequestActor       `json:"resolved_by,omitempty"`
	CreatedOn  string                 `json:"created_on,omitempty"`
	UpdatedOn  string                 `json:"updated_on,omitempty"`
	Links      PullRequestTaskLinks   `json:"links,omitempty"`
	Comment    *PullRequestComment    `json:"comment,omitempty"`
}

type PullRequestTaskContent struct {
	Raw    string `json:"raw,omitempty"`
	Markup string `json:"markup,omitempty"`
	HTML   string `json:"html,omitempty"`
}

type PullRequestTaskLinks struct {
	Self Link `json:"self,omitempty"`
	HTML Link `json:"html,omitempty"`
}

type pullRequestTaskListResponse struct {
	Values []PullRequestTask `json:"values"`
	Next   string            `json:"next,omitempty"`
}

type mergeTaskStatusResponse struct {
	TaskStatus  string      `json:"task_status"`
	MergeResult PullRequest `json:"merge_result,omitempty"`
}

const (
	mergeTaskStatusPending = "PENDING"
	mergeTaskStatusSuccess = "SUCCESS"

	defaultMergePollInterval = time.Second
	defaultMergePollTimeout  = 2 * time.Minute
)

func (c *Client) ListPullRequests(ctx context.Context, workspace, repoSlug string, options ListPullRequestsOptions) ([]PullRequest, error) {
	if options.Limit <= 0 {
		options.Limit = 20
	}

	path, err := listPullRequestsPath(workspace, repoSlug, options)
	if err != nil {
		return nil, err
	}

	var all []PullRequest
	nextPath := path

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestListResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pull request list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) GetPullRequest(ctx context.Context, workspace, repoSlug string, id int) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequest{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, workspace, repoSlug string, options CreatePullRequestOptions) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if options.Title == "" {
		return PullRequest{}, fmt.Errorf("pull request title is required")
	}
	if options.SourceBranch == "" {
		return PullRequest{}, fmt.Errorf("source branch is required")
	}
	if options.DestinationBranch == "" {
		return PullRequest{}, fmt.Errorf("destination branch is required")
	}

	if options.ReuseExisting {
		existing, err := c.ListPullRequests(ctx, workspace, repoSlug, ListPullRequestsOptions{
			State: "OPEN",
			Limit: 50,
		})
		if err != nil {
			return PullRequest{}, err
		}
		for _, pr := range existing {
			if pr.Title == options.Title && pr.Source.Branch.Name == options.SourceBranch && pr.Destination.Branch.Name == options.DestinationBranch {
				return pr, nil
			}
		}
	}

	body := map[string]any{
		"title": options.Title,
		"source": map[string]any{
			"branch": map[string]string{
				"name": options.SourceBranch,
			},
		},
		"destination": map[string]any{
			"branch": map[string]string{
				"name": options.DestinationBranch,
			},
		},
	}
	if options.Description != "" {
		body["description"] = options.Description
	}
	if options.CloseSourceBranch {
		body["close_source_branch"] = true
	}
	if options.Draft {
		body["draft"] = true
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequest{}, fmt.Errorf("marshal create pull request request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests", url.PathEscape(workspace), url.PathEscape(repoSlug))
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode created pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) MergePullRequest(ctx context.Context, workspace, repoSlug string, id int, options MergePullRequestOptions) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequest{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	body := map[string]any{
		"type": "pullrequest",
	}
	if strings.TrimSpace(options.Message) != "" {
		body["message"] = strings.TrimSpace(options.Message)
	}
	if options.CloseSourceBranch {
		body["close_source_branch"] = true
	}
	if strings.TrimSpace(options.MergeStrategy) != "" {
		body["merge_strategy"] = strings.TrimSpace(options.MergeStrategy)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequest{}, fmt.Errorf("marshal merge pull request request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/merge?async=true", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusAccepted:
		location := strings.TrimSpace(resp.Header.Get("Location"))
		if location == "" {
			return PullRequest{}, fmt.Errorf("bitbucket API returned 202 Accepted without a merge task location")
		}
		return c.waitForMergeTask(ctx, location, options.PollInterval, options.PollTimeout)
	default:
		if err := requireSuccess(resp); err != nil {
			return PullRequest{}, err
		}
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode merged pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) GetPullRequestPatch(ctx context.Context, workspace, repoSlug string, id int) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return "", fmt.Errorf("pull request ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/patch", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, map[string]string{
		"Accept": "text/plain",
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read pull request patch: %w", err)
	}

	return string(body), nil
}

func (c *Client) ListPullRequestDiffStats(ctx context.Context, workspace, repoSlug string, id int) ([]PullRequestDiffStat, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return nil, fmt.Errorf("pull request ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/diffstat?pagelen=100", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	nextPath := path
	var all []PullRequestDiffStat

	for nextPath != "" {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestDiffStatResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pull request diffstat: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	return all, nil
}

func (c *Client) CreatePullRequestComment(ctx context.Context, workspace, repoSlug string, id int, body string) (PullRequestComment, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestComment{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequestComment{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if strings.TrimSpace(body) == "" {
		return PullRequestComment{}, fmt.Errorf("comment body is required")
	}

	payload, err := json.Marshal(map[string]any{
		"content": map[string]string{
			"raw": body,
		},
	})
	if err != nil {
		return PullRequestComment{}, fmt.Errorf("marshal pull request comment request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequestComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestComment{}, err
	}

	var comment PullRequestComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return PullRequestComment{}, fmt.Errorf("decode pull request comment: %w", err)
	}

	return comment, nil
}

func (c *Client) GetPullRequestComment(ctx context.Context, workspace, repoSlug string, pullRequestID, commentID int) (PullRequestComment, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestComment{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestComment{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if commentID <= 0 {
		return PullRequestComment{}, fmt.Errorf("pull request comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, commentID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PullRequestComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestComment{}, err
	}

	var comment PullRequestComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return PullRequestComment{}, fmt.Errorf("decode pull request comment: %w", err)
	}

	return comment, nil
}

func (c *Client) UpdatePullRequestComment(ctx context.Context, workspace, repoSlug string, pullRequestID, commentID int, body string) (PullRequestComment, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestComment{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestComment{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if commentID <= 0 {
		return PullRequestComment{}, fmt.Errorf("pull request comment ID must be greater than zero")
	}
	if strings.TrimSpace(body) == "" {
		return PullRequestComment{}, fmt.Errorf("comment body is required")
	}

	payload, err := json.Marshal(map[string]any{
		"content": map[string]string{
			"raw": body,
		},
	})
	if err != nil {
		return PullRequestComment{}, fmt.Errorf("marshal pull request comment request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, commentID)
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return PullRequestComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestComment{}, err
	}

	var comment PullRequestComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return PullRequestComment{}, fmt.Errorf("decode updated pull request comment: %w", err)
	}

	return comment, nil
}

func (c *Client) DeletePullRequestComment(ctx context.Context, workspace, repoSlug string, pullRequestID, commentID int) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return fmt.Errorf("pull request ID must be greater than zero")
	}
	if commentID <= 0 {
		return fmt.Errorf("pull request comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, commentID)
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}

func (c *Client) ResolvePullRequestComment(ctx context.Context, workspace, repoSlug string, pullRequestID, commentID int) (PullRequestCommentResolve, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestCommentResolve{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestCommentResolve{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if commentID <= 0 {
		return PullRequestCommentResolve{}, fmt.Errorf("pull request comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, commentID)
	resp, err := c.Do(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return PullRequestCommentResolve{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestCommentResolve{}, err
	}

	var resolution PullRequestCommentResolve
	if err := json.NewDecoder(resp.Body).Decode(&resolution); err != nil {
		return PullRequestCommentResolve{}, fmt.Errorf("decode pull request comment resolution: %w", err)
	}

	return resolution, nil
}

func (c *Client) ReopenPullRequestComment(ctx context.Context, workspace, repoSlug string, pullRequestID, commentID int) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return fmt.Errorf("pull request ID must be greater than zero")
	}
	if commentID <= 0 {
		return fmt.Errorf("pull request comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/comments/%d/resolve", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, commentID)
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}

func (c *Client) ListPullRequestTasks(ctx context.Context, workspace, repoSlug string, pullRequestID int, options ListPullRequestTasksOptions) ([]PullRequestTask, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return nil, fmt.Errorf("pull request ID must be greater than zero")
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}

	path, err := listPullRequestTasksPath(workspace, repoSlug, pullRequestID, options)
	if err != nil {
		return nil, err
	}

	nextPath := path
	all := make([]PullRequestTask, 0)

	for nextPath != "" && len(all) < options.Limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page pullRequestTaskListResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode pull request task list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > options.Limit {
		all = all[:options.Limit]
	}

	return all, nil
}

func (c *Client) GetPullRequestTask(ctx context.Context, workspace, repoSlug string, pullRequestID, taskID int) (PullRequestTask, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestTask{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestTask{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if taskID <= 0 {
		return PullRequestTask{}, fmt.Errorf("pull request task ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/tasks/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, taskID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return PullRequestTask{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestTask{}, err
	}

	var task PullRequestTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return PullRequestTask{}, fmt.Errorf("decode pull request task: %w", err)
	}

	return task, nil
}

func (c *Client) CreatePullRequestTask(ctx context.Context, workspace, repoSlug string, pullRequestID int, options CreatePullRequestTaskOptions) (PullRequestTask, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestTask{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestTask{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if strings.TrimSpace(options.Body) == "" {
		return PullRequestTask{}, fmt.Errorf("task body is required")
	}
	if options.CommentID < 0 {
		return PullRequestTask{}, fmt.Errorf("pull request comment ID must be greater than zero")
	}

	body := map[string]any{
		"content": map[string]string{
			"raw": strings.TrimSpace(options.Body),
		},
	}
	if options.CommentID > 0 {
		body["comment"] = map[string]int{"id": options.CommentID}
	}
	if options.Pending {
		body["pending"] = true
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequestTask{}, fmt.Errorf("marshal pull request task request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/tasks", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return PullRequestTask{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestTask{}, err
	}

	var task PullRequestTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return PullRequestTask{}, fmt.Errorf("decode created pull request task: %w", err)
	}

	return task, nil
}

func (c *Client) UpdatePullRequestTask(ctx context.Context, workspace, repoSlug string, pullRequestID, taskID int, options UpdatePullRequestTaskOptions) (PullRequestTask, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequestTask{}, fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return PullRequestTask{}, fmt.Errorf("pull request ID must be greater than zero")
	}
	if taskID <= 0 {
		return PullRequestTask{}, fmt.Errorf("pull request task ID must be greater than zero")
	}
	if strings.TrimSpace(options.Body) == "" && strings.TrimSpace(options.State) == "" {
		return PullRequestTask{}, fmt.Errorf("task update requires --body and/or --state")
	}

	body := map[string]any{}
	if strings.TrimSpace(options.Body) != "" {
		body["content"] = map[string]string{
			"raw": strings.TrimSpace(options.Body),
		}
	}
	if strings.TrimSpace(options.State) != "" {
		body["state"] = strings.TrimSpace(options.State)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return PullRequestTask{}, fmt.Errorf("marshal pull request task update request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/tasks/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, taskID)
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return PullRequestTask{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequestTask{}, err
	}

	var task PullRequestTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return PullRequestTask{}, fmt.Errorf("decode updated pull request task: %w", err)
	}

	return task, nil
}

func (c *Client) DeletePullRequestTask(ctx context.Context, workspace, repoSlug string, pullRequestID, taskID int) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return fmt.Errorf("pull request ID must be greater than zero")
	}
	if taskID <= 0 {
		return fmt.Errorf("pull request task ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/tasks/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, taskID)
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return requireSuccess(resp)
}

func (c *Client) DeclinePullRequest(ctx context.Context, workspace, repoSlug string, id int) (PullRequest, error) {
	if workspace == "" || repoSlug == "" {
		return PullRequest{}, fmt.Errorf("workspace and repository are required")
	}
	if id <= 0 {
		return PullRequest{}, fmt.Errorf("pull request ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/decline", url.PathEscape(workspace), url.PathEscape(repoSlug), id)
	resp, err := c.Do(ctx, http.MethodPost, path, nil, nil)
	if err != nil {
		return PullRequest{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PullRequest{}, fmt.Errorf("decode declined pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) waitForMergeTask(ctx context.Context, taskURL string, interval, timeout time.Duration) (PullRequest, error) {
	if interval <= 0 {
		interval = defaultMergePollInterval
	}
	if timeout <= 0 {
		timeout = defaultMergePollTimeout
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		resp, err := c.Do(pollCtx, http.MethodGet, taskURL, nil, nil)
		if err != nil {
			return PullRequest{}, err
		}

		var status mergeTaskStatusResponse
		func() {
			defer resp.Body.Close()
			if err != nil {
				return
			}
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&status)
		}()
		if err != nil {
			return PullRequest{}, fmt.Errorf("check merge task status: %w", err)
		}

		switch status.TaskStatus {
		case mergeTaskStatusSuccess:
			if status.MergeResult.ID == 0 {
				return PullRequest{}, fmt.Errorf("merge task completed without a merged pull request payload")
			}
			return status.MergeResult, nil
		case "", mergeTaskStatusPending:
			if status.MergeResult.ID != 0 {
				return status.MergeResult, nil
			}
		default:
			return PullRequest{}, fmt.Errorf("merge task returned unexpected status %q", status.TaskStatus)
		}

		select {
		case <-pollCtx.Done():
			return PullRequest{}, fmt.Errorf("timed out waiting for merge task: %w", pollCtx.Err())
		case <-time.After(interval):
		}
	}
}

func listPullRequestsPath(workspace, repoSlug string, options ListPullRequestsOptions) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(options.Limit))
	if options.State != "" && options.State != "ALL" {
		values.Set("state", options.State)
	}
	if options.Query != "" {
		values.Set("q", options.Query)
	}
	if options.Sort != "" {
		values.Set("sort", options.Sort)
	}

	return fmt.Sprintf("/repositories/%s/%s/pullrequests?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), values.Encode()), nil
}

func listPullRequestTasksPath(workspace, repoSlug string, pullRequestID int, options ListPullRequestTasksOptions) (string, error) {
	if workspace == "" || repoSlug == "" {
		return "", fmt.Errorf("workspace and repository are required")
	}
	if pullRequestID <= 0 {
		return "", fmt.Errorf("pull request ID must be greater than zero")
	}

	values := url.Values{}
	pagelen := options.Limit
	if pagelen <= 0 {
		pagelen = 20
	}
	if pagelen > 100 {
		pagelen = 100
	}
	values.Set("pagelen", strconv.Itoa(pagelen))

	query := strings.TrimSpace(options.Query)
	if state := strings.TrimSpace(options.State); state != "" && !strings.EqualFold(state, "ALL") {
		stateQuery := fmt.Sprintf("state=\"%s\"", state)
		if query == "" {
			query = stateQuery
		} else {
			query = fmt.Sprintf("(%s) AND %s", query, stateQuery)
		}
	}
	if query != "" {
		values.Set("q", query)
	}
	if sort := strings.TrimSpace(options.Sort); sort != "" {
		values.Set("sort", sort)
	}

	return fmt.Sprintf("/repositories/%s/%s/pullrequests/%d/tasks?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), pullRequestID, values.Encode()), nil
}
