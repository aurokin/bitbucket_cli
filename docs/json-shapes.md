# JSON Shapes

Representative JSON payload shapes for common commands.

These examples are based on the current command payload structs and Bitbucket Cloud response models. Field order in the actual output may differ, and omitted fields depend on the selected command, flags, and data returned by Bitbucket.

Use [automation.md](./automation.md) for deterministic command patterns and [cli-reference.md](./cli-reference.md) for the full command surface.

## Repository View

Command:

```bash
bb repo view --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "host": "bitbucket.org",
  "workspace": "OhBizzle",
  "repo": "bb-cli-integration-primary",
  "name": "bb-cli-integration-primary",
  "full_name": "OhBizzle/bb-cli-integration-primary",
  "description": "Integration fixture repository",
  "private": true,
  "project_key": "BBCLI",
  "project_name": "Bitbucket CLI",
  "main_branch": "main",
  "html_url": "https://bitbucket.org/OhBizzle/bb-cli-integration-primary",
  "https_clone": "https://bitbucket.org/OhBizzle/bb-cli-integration-primary.git",
  "ssh_clone": "git@bitbucket.org:OhBizzle/bb-cli-integration-primary.git",
  "remote": "origin",
  "local_clone_url": "git@bitbucket.org:OhBizzle/bb-cli-integration-primary.git",
  "root": "/path/to/checkout"
}
```

## Repository Clone And Delete

Commands:

```bash
bb repo clone OhBizzle/bb-cli-integration-primary /tmp/bb-cli-integration-primary --json '*'
bb --no-prompt repo delete OhBizzle/example-repo --yes --json '*'
```

Representative shapes:

```json
{
  "host": "bitbucket.org",
  "workspace": "OhBizzle",
  "repo": "bb-cli-integration-primary",
  "name": "bb-cli-integration-primary",
  "directory": "/tmp/bb-cli-integration-primary",
  "clone_url": "https://bitbucket.org/OhBizzle/bb-cli-integration-primary.git"
}
```

```json
{
  "host": "bitbucket.org",
  "workspace": "OhBizzle",
  "repo": "example-repo",
  "name": "example-repo",
  "deleted": true
}
```

## Pull Request List And View

Commands:

```bash
bb pr list --repo OhBizzle/bb-cli-integration-primary --json '*'
bb pr view 1 --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative list item shape:

```json
{
  "id": 1,
  "title": "bb cli integration fixture pull request",
  "description": "Integration fixture PR",
  "state": "OPEN",
  "author": {
    "display_name": "Hunter Sadler",
    "account_id": "5afc40f1a496f735aacc815e",
    "nickname": "OhBizzle"
  },
  "reviewers": [],
  "participants": [],
  "source": {
    "branch": {
      "name": "feature"
    }
  },
  "destination": {
    "branch": {
      "name": "main"
    }
  },
  "close_source_branch": false,
  "draft": false,
  "created_on": "2026-03-11T00:00:00.000000+00:00",
  "updated_on": "2026-03-11T00:00:00.000000+00:00",
  "links": {
    "html": {
      "href": "https://bitbucket.org/OhBizzle/bb-cli-integration-primary/pull-requests/1"
    }
  }
}
```

## Pull Request Status

Command:

```bash
bb pr status --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "host": "bitbucket.org",
  "workspace": "OhBizzle",
  "repo": "bb-cli-integration-primary",
  "current_user": {
    "account_id": "5afc40f1a496f735aacc815e",
    "display_name": "Hunter Sadler",
    "nickname": "OhBizzle"
  },
  "current_branch_name": "feature",
  "current_branch": {
    "id": 2,
    "title": "bb cli create command pull request",
    "state": "OPEN"
  },
  "created": [],
  "review_requested": []
}
```

## Pull Request Diff

Command:

```bash
bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "host": "bitbucket.org",
  "workspace": "OhBizzle",
  "repo": "bb-cli-integration-primary",
  "id": 1,
  "title": "bb cli integration fixture pull request",
  "patch": "diff --git a/file.txt b/file.txt\n...",
  "stats": [
    {
      "type": "diffstat",
      "status": "modified",
      "lines_added": 5,
      "lines_removed": 1,
      "old": {
        "path": "file.txt"
      },
      "new": {
        "path": "file.txt"
      }
    }
  ]
}
```

## Issue List And View

Commands:

```bash
bb issue list --repo OhBizzle/bb-cli-integration-issues --json '*'
bb issue view 1 --repo OhBizzle/bb-cli-integration-issues --json '*'
```

Representative issue shape:

```json
{
  "id": 1,
  "title": "Broken flow",
  "state": "new",
  "kind": "bug",
  "priority": "major",
  "content": {
    "raw": "Needs investigation."
  },
  "reporter": {
    "display_name": "Hunter Sadler",
    "account_id": "5afc40f1a496f735aacc815e",
    "nickname": "OhBizzle"
  },
  "assignee": {},
  "created_on": "2026-03-11T00:00:00.000000+00:00",
  "updated_on": "2026-03-11T00:00:00.000000+00:00",
  "links": {
    "html": {
      "href": "https://bitbucket.org/OhBizzle/bb-cli-integration-issues/issues/1"
    }
  }
}
```

## Search And Cross-Repository Status

Commands:

```bash
bb search repos bb-cli --workspace OhBizzle --json '*'
bb status --workspace OhBizzle --json '*'
```

Representative `bb status` shape:

```json
{
  "user": "Hunter Sadler",
  "workspaces": [
    "OhBizzle"
  ],
  "repositories_scanned": 3,
  "repo_limit_per_workspace": 100,
  "item_limit_per_section": 20,
  "authored_prs_total": 2,
  "review_requested_prs_total": 0,
  "your_issues_total": 1,
  "repositories_without_issue_tracker": 2,
  "workspaces_at_repo_limit": [],
  "warnings": [
    "2 repositories do not have issue tracking enabled"
  ],
  "authored_prs": [
    {
      "workspace": "OhBizzle",
      "repo": "bb-cli-integration-primary",
      "pull_request": {
        "id": 1,
        "title": "bb cli integration fixture pull request",
        "state": "OPEN"
      }
    }
  ],
  "review_requested_prs": [],
  "your_issues": [
    {
      "workspace": "OhBizzle",
      "repo": "bb-cli-integration-issues",
      "issue": {
        "id": 1,
        "title": "Broken flow",
        "state": "new"
      }
    }
  ]
}
```
