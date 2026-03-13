package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type IssueAttachment struct {
	Name  string               `json:"name,omitempty"`
	Links IssueAttachmentLinks `json:"links,omitempty"`
}

type IssueAttachmentLinks struct {
	Self IssueAttachmentLink `json:"self,omitempty"`
	HTML IssueAttachmentLink `json:"html,omitempty"`
}

type IssueAttachmentLink struct {
	Href string `json:"href,omitempty"`
}

func (l *IssueAttachmentLink) UnmarshalJSON(data []byte) error {
	var single struct {
		Href string `json:"href"`
	}
	if err := json.Unmarshal(data, &single); err == nil && single.Href != "" {
		l.Href = single.Href
		return nil
	}

	var multi struct {
		Href []string `json:"href"`
	}
	if err := json.Unmarshal(data, &multi); err == nil {
		if len(multi.Href) > 0 {
			l.Href = multi.Href[0]
		}
		return nil
	}

	return fmt.Errorf("decode issue attachment link")
}

type issueAttachmentListResponse struct {
	Values []IssueAttachment `json:"values"`
	Next   string            `json:"next,omitempty"`
}

func (c *Client) ListIssueAttachments(ctx context.Context, workspace, repoSlug string, issueID, limit int) ([]IssueAttachment, error) {
	if workspace == "" || repoSlug == "" {
		return nil, fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return nil, fmt.Errorf("issue ID must be greater than zero")
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	values := url.Values{}
	values.Set("pagelen", strconv.Itoa(limit))
	nextPath := fmt.Sprintf("/repositories/%s/%s/issues/%d/attachments?%s", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID, values.Encode())
	all := make([]IssueAttachment, 0)

	for nextPath != "" && len(all) < limit {
		resp, err := c.Do(ctx, http.MethodGet, nextPath, nil, nil)
		if err != nil {
			return nil, err
		}
		var page issueAttachmentListResponse
		func() {
			defer resp.Body.Close()
			err = requireSuccess(resp)
			if err != nil {
				return
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
		}()
		if err != nil {
			return nil, fmt.Errorf("decode issue attachment list: %w", err)
		}
		all = append(all, page.Values...)
		nextPath = page.Next
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (c *Client) UploadIssueAttachments(ctx context.Context, workspace, repoSlug string, issueID int, paths []string) error {
	if workspace == "" || repoSlug == "" {
		return fmt.Errorf("workspace and repository are required")
	}
	if issueID <= 0 {
		return fmt.Errorf("issue ID must be greater than zero")
	}
	if len(paths) == 0 {
		return fmt.Errorf("at least one issue attachment path is required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, filePath := range paths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read issue attachment %q: %w", filePath, err)
		}
		part, err := writer.CreateFormFile("files", filepath.Base(filePath))
		if err != nil {
			return fmt.Errorf("create multipart part for %q: %w", filePath, err)
		}
		if _, err := part.Write(data); err != nil {
			return fmt.Errorf("write multipart part for %q: %w", filePath, err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart body: %w", err)
	}

	path := fmt.Sprintf("/repositories/%s/%s/issues/%d/attachments", url.PathEscape(workspace), url.PathEscape(repoSlug), issueID)
	resp, err := c.Do(ctx, http.MethodPost, path, body.Bytes(), map[string]string{"Content-Type": writer.FormDataContentType()})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}
