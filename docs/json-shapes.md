# JSON Shapes

Representative JSON payload shapes for common commands.

Generated from the current payload structs and Bitbucket response models. Field order follows the Go structs. Omitted fields in live output still depend on the selected command, flags, and Bitbucket data.

Use [automation.md](./automation.md) for deterministic command patterns and [cli-reference.md](./cli-reference.md) for the full command surface.

## Repository View

Representative shape for the repository view payload.

Command:

```bash
bb repo view --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "description": "Example text",
  "full_name": "workspace-slug/repo-slug",
  "host": "bitbucket.org",
  "html_url": "https://bitbucket.org/workspace-slug/repo-slug",
  "https_clone": "https://bitbucket.org/workspace-slug/repo-slug.git",
  "local_clone_url": "git@bitbucket.org:workspace-slug/repo-slug.git",
  "main_branch": "main",
  "name": "Example Name",
  "private": true,
  "project_key": "BBCLI",
  "project_name": "Bitbucket CLI",
  "remote": "origin",
  "repo": "repo-slug",
  "root": "/path/to/repo",
  "ssh_clone": "git@bitbucket.org:workspace-slug/repo-slug.git",
  "workspace": "workspace-slug"
}
```

## Repository Clone And Delete

Representative shapes for repository clone and delete payloads.

Commands:

```bash
bb repo clone OhBizzle/bb-cli-integration-primary /tmp/bb-cli-integration-primary --json '*'
bb --no-prompt repo delete OhBizzle/example-repo --yes --json '*'
```

Representative shapes:

```json
{
  "clone_url": "https://bitbucket.org/workspace-slug/repo-slug.git",
  "directory": "/path/to/repo",
  "host": "bitbucket.org",
  "name": "Example Name",
  "repo": "repo-slug",
  "workspace": "workspace-slug"
}
```

```json
{
  "deleted": true,
  "host": "bitbucket.org",
  "name": "Example Name",
  "repo": "repo-slug",
  "workspace": "workspace-slug"
}
```

## Browse

Representative shape for browse payloads when printing or emitting JSON instead of opening the browser.

Commands:

```bash
bb browse --repo OhBizzle/bb-cli-integration-primary --no-browser --json '*'
bb browse README.md:12 --repo OhBizzle/bb-cli-integration-primary --no-browser --json '*'
bb browse --pr 1 --repo OhBizzle/bb-cli-integration-primary --no-browser --json '*'
```

Representative shape:

```json
{
  "commit": "\u003ccommit\u003e",
  "host": "bitbucket.org",
  "issue": 1,
  "line": 1,
  "opened": true,
  "path": "file.txt",
  "pr": 1,
  "ref": "\u003cref\u003e",
  "repo": "repo-slug",
  "type": "commit_file",
  "url": "\u003curl\u003e",
  "workspace": "workspace-slug"
}
```

## Pipeline List And View

Representative shapes for pipeline list items plus pipeline log, stop, and view payloads.

Commands:

```bash
bb pipeline list --repo OhBizzle/bb-cli-integration-pipelines --json '*'
bb pipeline log 1 --repo OhBizzle/bb-cli-integration-pipelines --step '{step-uuid}' --json '*'
bb --no-prompt pipeline stop 1 --repo OhBizzle/bb-cli-integration-pipelines --yes --json '*'
bb pipeline view 1 --repo OhBizzle/bb-cli-integration-pipelines --json '*'
```

Representative shapes:

```json
{
  "build_number": 1,
  "completed_on": "\u003ccompleted-on\u003e",
  "created_on": "2026-03-11T00:00:00Z",
  "creator": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "state": {
    "name": "Example Name",
    "result": {
      "name": "Example Name",
      "type": "commit_file"
    },
    "stage": {
      "name": "Example Name",
      "type": "commit_file"
    },
    "type": "commit_file"
  },
  "target": {
    "commit": {
      "hash": "abc123def456"
    },
    "ref_name": "\u003cref-name\u003e",
    "ref_type": "\u003cref-type\u003e",
    "selector": {
      "pattern": "\u003cpattern\u003e",
      "type": "commit_file"
    },
    "type": "commit_file"
  },
  "uuid": "{uuid}"
}
```

```json
{
  "host": "bitbucket.org",
  "log": "\u003clog\u003e",
  "pipeline": {
    "build_number": 1,
    "completed_on": "\u003ccompleted-on\u003e",
    "created_on": "2026-03-11T00:00:00Z",
    "creator": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "state": {
      "name": "Example Name",
      "result": {
        "name": "Example Name"
      },
      "stage": {
        "name": "Example Name"
      },
      "type": "commit_file"
    },
    "target": {
      "commit": {},
      "ref_name": "\u003cref-name\u003e",
      "ref_type": "\u003cref-type\u003e",
      "selector": {},
      "type": "commit_file"
    },
    "uuid": "{uuid}"
  },
  "repo": "repo-slug",
  "step": {
    "completed_on": "\u003ccompleted-on\u003e",
    "name": "Example Name",
    "started_on": "\u003cstarted-on\u003e",
    "state": {
      "name": "Example Name",
      "result": {
        "name": "Example Name"
      },
      "stage": {
        "name": "Example Name"
      },
      "type": "commit_file"
    },
    "uuid": "{uuid}"
  },
  "workspace": "workspace-slug"
}
```

```json
{
  "host": "bitbucket.org",
  "pipeline": {
    "build_number": 1,
    "completed_on": "\u003ccompleted-on\u003e",
    "created_on": "2026-03-11T00:00:00Z",
    "creator": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "state": {
      "name": "Example Name",
      "result": {
        "name": "Example Name"
      },
      "stage": {
        "name": "Example Name"
      },
      "type": "commit_file"
    },
    "target": {
      "commit": {},
      "ref_name": "\u003cref-name\u003e",
      "ref_type": "\u003cref-type\u003e",
      "selector": {},
      "type": "commit_file"
    },
    "uuid": "{uuid}"
  },
  "repo": "repo-slug",
  "stopped": true,
  "workspace": "workspace-slug"
}
```

```json
{
  "host": "bitbucket.org",
  "pipeline": {
    "build_number": 1,
    "completed_on": "\u003ccompleted-on\u003e",
    "created_on": "2026-03-11T00:00:00Z",
    "creator": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "state": {
      "name": "Example Name",
      "result": {
        "name": "Example Name"
      },
      "stage": {
        "name": "Example Name"
      },
      "type": "commit_file"
    },
    "target": {
      "commit": {},
      "ref_name": "\u003cref-name\u003e",
      "ref_type": "\u003cref-type\u003e",
      "selector": {},
      "type": "commit_file"
    },
    "uuid": "{uuid}"
  },
  "repo": "repo-slug",
  "steps": [
    {
      "completed_on": "\u003ccompleted-on\u003e",
      "name": "Example Name",
      "started_on": "\u003cstarted-on\u003e",
      "state": {
        "name": "Example Name"
      },
      "uuid": "{uuid}"
    }
  ],
  "workspace": "workspace-slug"
}
```

## Pull Request List And View

Representative shape for pull request list items and the pull request view payload.

Commands:

```bash
bb pr list --repo OhBizzle/bb-cli-integration-primary --json '*'
bb pr view 1 --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "author": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "close_source_branch": true,
  "comment_count": 2,
  "created_on": "2026-03-11T00:00:00Z",
  "description": "Example text",
  "destination": {
    "branch": {
      "default_merge_strategy": "merge_commit",
      "merge_strategies": [
        "merge_commit"
      ],
      "name": "main"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "full_name": "workspace-slug/repo-slug",
      "name": "Example Name"
    }
  },
  "draft": true,
  "id": 1,
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "merge_commit": {
    "hash": "abc123def456"
  },
  "participants": [
    {
      "approved": true,
      "role": "REVIEWER",
      "state": "OPEN",
      "user": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      }
    }
  ],
  "queued": true,
  "reviewers": [
    {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    }
  ],
  "source": {
    "branch": {
      "default_merge_strategy": "merge_commit",
      "merge_strategies": [
        "merge_commit"
      ],
      "name": "main"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "full_name": "workspace-slug/repo-slug",
      "name": "Example Name"
    }
  },
  "state": "OPEN",
  "task_count": 2,
  "title": "Example title",
  "updated_on": "2026-03-11T00:00:00Z"
}
```

## Pull Request Status

Representative shape for the pull request status payload.

Command:

```bash
bb pr status --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "created": [
    {
      "author": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      },
      "description": "Example text",
      "id": 1,
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      },
      "source": {
        "branch": {
          "name": "main"
        },
        "repository": {
          "name": "Example Name"
        }
      },
      "state": "OPEN",
      "title": "Example title"
    }
  ],
  "current_branch": {
    "author": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "description": "Example text",
    "id": 1,
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "source": {
      "branch": {
        "name": "main"
      },
      "repository": {
        "name": "Example Name"
      }
    },
    "state": "OPEN",
    "title": "Example title"
  },
  "current_branch_error": "\u003ccurrent-branch-error\u003e",
  "current_branch_name": "\u003ccurrent-branch-name\u003e",
  "current_user": {
    "account_id": "account-id",
    "display_name": "Example User",
    "username": "user@example.com",
    "uuid": "{uuid}"
  },
  "host": "bitbucket.org",
  "repo": "repo-slug",
  "review_requested": [
    {
      "author": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      },
      "description": "Example text",
      "id": 1,
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      },
      "source": {
        "branch": {
          "name": "main"
        },
        "repository": {
          "name": "Example Name"
        }
      },
      "state": "OPEN",
      "title": "Example title"
    }
  ],
  "workspace": "workspace-slug"
}
```

## Pull Request Diff

Representative shape for the pull request diff payload.

Command:

```bash
bb pr diff 1 --repo OhBizzle/bb-cli-integration-primary --json '*'
```

Representative shape:

```json
{
  "host": "bitbucket.org",
  "id": 1,
  "patch": "diff --git a/file.txt b/file.txt\n...",
  "repo": "repo-slug",
  "stats": [
    {
      "lines_added": 1,
      "lines_removed": 1,
      "new": {
        "path": "file.txt"
      },
      "old": {
        "path": "file.txt"
      },
      "status": "modified"
    }
  ],
  "title": "Example title",
  "workspace": "workspace-slug"
}
```

## Issue List And View

Representative shape for issue list items and the issue view payload.

Commands:

```bash
bb issue list --repo OhBizzle/bb-cli-integration-issues --json '*'
bb issue view 1 --repo OhBizzle/bb-cli-integration-issues --json '*'
```

Representative shape:

```json
{
  "assignee": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "content": {
    "raw": "Example text"
  },
  "created_on": "2026-03-11T00:00:00Z",
  "id": 1,
  "kind": "\u003ckind\u003e",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "priority": "\u003cpriority\u003e",
  "reporter": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "state": "OPEN",
  "title": "Example title",
  "updated_on": "2026-03-11T00:00:00Z"
}
```

## Search Results

Representative shapes for repository, pull request, and issue search results.

Commands:

```bash
bb search repos bb-cli --workspace OhBizzle --json '*'
bb search prs fixture --repo OhBizzle/bb-cli-integration-primary --json '*'
bb search issues broken --repo OhBizzle/bb-cli-integration-issues --json '*'
```

Representative shapes:

```json
{
  "description": "Example text",
  "full_name": "workspace-slug/repo-slug",
  "is_private": true,
  "links": {
    "clone": [
      {
        "href": "https://bitbucket.org/workspace-slug/repo-slug",
        "name": "Example Name"
      }
    ],
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "mainbranch": {
    "name": "Example Name"
  },
  "name": "Example Name",
  "project": {
    "key": "BBCLI",
    "name": "BBCLI"
  },
  "slug": "\u003cslug\u003e",
  "updated_on": "2026-03-11T00:00:00Z"
}
```

```json
{
  "author": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "close_source_branch": true,
  "comment_count": 2,
  "created_on": "2026-03-11T00:00:00Z",
  "description": "Example text",
  "destination": {
    "branch": {
      "default_merge_strategy": "merge_commit",
      "merge_strategies": [
        "merge_commit"
      ],
      "name": "main"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "full_name": "workspace-slug/repo-slug",
      "name": "Example Name"
    }
  },
  "draft": true,
  "id": 1,
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "merge_commit": {
    "hash": "abc123def456"
  },
  "participants": [
    {
      "approved": true,
      "role": "REVIEWER",
      "state": "OPEN",
      "user": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      }
    }
  ],
  "queued": true,
  "reviewers": [
    {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    }
  ],
  "source": {
    "branch": {
      "default_merge_strategy": "merge_commit",
      "merge_strategies": [
        "merge_commit"
      ],
      "name": "main"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "full_name": "workspace-slug/repo-slug",
      "name": "Example Name"
    }
  },
  "state": "OPEN",
  "task_count": 2,
  "title": "Example title",
  "updated_on": "2026-03-11T00:00:00Z"
}
```

```json
{
  "assignee": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "content": {
    "raw": "Example text"
  },
  "created_on": "2026-03-11T00:00:00Z",
  "id": 1,
  "kind": "\u003ckind\u003e",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  },
  "priority": "\u003cpriority\u003e",
  "reporter": {
    "account_id": "account-id",
    "display_name": "Example User",
    "nickname": "example-user"
  },
  "state": "OPEN",
  "title": "Example title",
  "updated_on": "2026-03-11T00:00:00Z"
}
```

## Cross-Repository Status

Representative shape for the cross-repository status payload.

Command:

```bash
bb status --workspace OhBizzle --json '*'
```

Representative shape:

```json
{
  "authored_prs": [
    {
      "pull_request": {
        "author": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        },
        "description": "Example text",
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "source": {
          "branch": {
            "name": "main"
          },
          "repository": {
            "name": "Example Name"
          }
        },
        "state": "OPEN",
        "title": "Example title"
      },
      "repo": "repo-slug",
      "workspace": "workspace-slug"
    }
  ],
  "authored_prs_total": 3,
  "item_limit_per_section": 20,
  "repo_limit_per_workspace": 20,
  "repositories_scanned": 4,
  "repositories_without_issue_tracker": 1,
  "review_requested_prs": [
    {
      "pull_request": {
        "author": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        },
        "description": "Example text",
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "source": {
          "branch": {
            "name": "main"
          },
          "repository": {
            "name": "Example Name"
          }
        },
        "state": "OPEN",
        "title": "Example title"
      },
      "repo": "repo-slug",
      "workspace": "workspace-slug"
    }
  ],
  "review_requested_prs_total": 3,
  "user": "Example User",
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspaces": [
    "\u003citem\u003e"
  ],
  "workspaces_at_repo_limit": [
    "\u003citem\u003e"
  ],
  "your_issues": [
    {
      "issue": {
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "state": "OPEN",
        "title": "Example title"
      },
      "repo": "repo-slug",
      "workspace": "workspace-slug"
    }
  ],
  "your_issues_total": 3
}
```
