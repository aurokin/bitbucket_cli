package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type IssueComment struct {
	ID        int          `json:"id"`
	Content   IssueContent `json:"content,omitempty"`
	User      IssueActor   `json:"user,omitempty"`
	CreatedOn string       `json:"created_on,omitempty"`
	UpdatedOn string       `json:"updated_on,omitempty"`
	Deleted   bool         `json:"deleted,omitempty"`
	Links     IssueLinks   `json:"links,omitempty"`
}

type issueCommentListResponse struct {
	Values []IssueComment `json:"values"`
	Next   string         `json:"next,omitempty"`
}

func (c *Client) ListIssueComments(ctx context.Context, workspace, repoSlug string, issueID, limit int) ([]IssueComment, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return nil, fmt.Errorf("issue ID must be greater than zero")
	}
	if limit <= 0 {
		limit = 20
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID, values.Encode())
	all := make([]IssueComment, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}

		var page issueCommentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode issue comment list: %w", err)
		}

		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) GetIssueComment(ctx context.Context, workspace, repoSlug string, issueID, commentID int) (IssueComment, error) {
	if workspace == "" || repoSlug == "" {
		return IssueComment{}, fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return IssueComment{}, fmt.Errorf("issue ID must be greater than zero")
	}
	if commentID <= 0 {
		return IssueComment{}, fmt.Errorf("issue comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID, commentID)
	resp, err := c.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return IssueComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return IssueComment{}, err
	}

	var comment IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return IssueComment{}, fmt.Errorf("decode issue comment: %w", err)
	}
	return comment, nil
}

func (c *Client) CreateIssueComment(ctx context.Context, workspace, repoSlug string, issueID int, body string) (IssueComment, error) {
	if workspace == "" || repoSlug == "" {
		return IssueComment{}, fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return IssueComment{}, fmt.Errorf("issue ID must be greater than zero")
	}
	if body == "" {
		return IssueComment{}, fmt.Errorf("issue comment body is required")
	}

	payload, err := json.Marshal(map[string]any{"content": map[string]string{"raw": body}})
	if err != nil {
		return IssueComment{}, fmt.Errorf("marshal create issue comment request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID)
	resp, err := c.Do(ctx, http.MethodPost, path, payload, nil)
	if err != nil {
		return IssueComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return IssueComment{}, err
	}

	var comment IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return IssueComment{}, fmt.Errorf("decode created issue comment: %w", err)
	}
	return comment, nil
}

func (c *Client) UpdateIssueComment(ctx context.Context, workspace, repoSlug string, issueID, commentID int, body string) (IssueComment, error) {
	if workspace == "" || repoSlug == "" {
		return IssueComment{}, fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return IssueComment{}, fmt.Errorf("issue ID must be greater than zero")
	}
	if commentID <= 0 {
		return IssueComment{}, fmt.Errorf("issue comment ID must be greater than zero")
	}
	if body == "" {
		return IssueComment{}, fmt.Errorf("issue comment body is required")
	}

	payload, err := json.Marshal(map[string]any{"content": map[string]string{"raw": body}})
	if err != nil {
		return IssueComment{}, fmt.Errorf("marshal update issue comment request: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID, commentID)
	resp, err := c.Do(ctx, http.MethodPut, path, payload, nil)
	if err != nil {
		return IssueComment{}, err
	}
	defer resp.Body.Close()

	if err := requireSuccess(resp); err != nil {
		return IssueComment{}, err
	}

	var comment IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return IssueComment{}, fmt.Errorf("decode updated issue comment: %w", err)
	}
	return comment, nil
}

func (c *Client) DeleteIssueComment(ctx context.Context, workspace, repoSlug string, issueID, commentID int) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return fmt.Errorf("issue ID must be greater than zero")
	}
	if commentID <= 0 {
		return fmt.Errorf("issue comment ID must be greater than zero")
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/comments/%d", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID, commentID)
	resp, err := c.Do(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}
