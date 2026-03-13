# JSON Shapes

Representative JSON payload shapes for common commands.

Generated from the current payload structs and Bitbucket response models. Field order follows the Go structs. Omitted fields in live output still depend on the selected command, flags, and Bitbucket data.

Use [automation.md](./automation.md) for deterministic command patterns and [cli-reference.md](./cli-reference.md) for the full command surface.

## Repository View

Representative shape for the repository view payload.

Command:

```bash
bb repo view --repo workspace-slug/repo-slug --json '*'
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
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspace": "workspace-slug"
}
```

## Repository Clone And Delete

Representative shapes for repository clone and delete payloads.

Commands:

```bash
bb repo clone workspace-slug/repo-slug /tmp/repo-slug --json '*'
bb --no-prompt repo delete workspace-slug/delete-repo-slug --yes --json '*'
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
bb browse --repo workspace-slug/repo-slug --no-browser --json '*'
bb browse README.md:12 --repo workspace-slug/repo-slug --no-browser --json '*'
bb browse --pr 1 --repo workspace-slug/repo-slug --no-browser --json '*'
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
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspace": "workspace-slug"
}
```

## Resolve

Representative shape for URL-to-entity resolution payloads used by agents and humans to normalize Bitbucket URLs.

Command:

```bash
bb resolve https://bitbucket.org/workspace-slug/repo-slug/pull-requests/7#comment-15 --json '*'
```

Representative shape:

```json
{
  "canonical_url": "\u003ccanonical-url\u003e",
  "comment": 1,
  "commit": "\u003ccommit\u003e",
  "host": "bitbucket.org",
  "issue": 1,
  "line": 1,
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
bb pipeline list --repo workspace-slug/pipelines-repo-slug --json '*'
bb pipeline log 1 --repo workspace-slug/pipelines-repo-slug --step '{step-uuid}' --json '*'
bb --no-prompt pipeline stop 1 --repo workspace-slug/pipelines-repo-slug --yes --json '*'
bb pipeline view 1 --repo workspace-slug/pipelines-repo-slug --json '*'
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
  "warnings": [
    "\u003citem\u003e"
  ],
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
  "warnings": [
    "\u003citem\u003e"
  ],
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
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspace": "workspace-slug"
}
```

## Pull Request List And View

Representative shape for pull request list items and the pull request view payload.

Commands:

```bash
bb pr list --repo workspace-slug/repo-slug --json '*'
bb pr view 1 --repo workspace-slug/repo-slug --json '*'
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
      "participated_on": "\u003cparticipated-on\u003e",
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

## Pull Request Comment View And Resolution

Representative shape for pull request comment detail, edit, delete, resolve, and reopen payloads.

Commands:

```bash
bb pr comment view https://bitbucket.org/workspace-slug/repo-slug/pull-requests/1#comment-15 --json '*'
bb pr comment resolve 15 --pr 1 --repo workspace-slug/repo-slug --json '*'
```

Representative shape:

```json
{
  "action": "\u003caction\u003e",
  "comment": {
    "content": {
      "raw": "Example text"
    },
    "created_on": "2026-03-11T00:00:00Z",
    "deleted": true,
    "id": 1,
    "inline": {
      "from": 1,
      "path": "file.txt",
      "start_from": 1,
      "start_to": 1,
      "to": 1
    },
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "parent": {
      "content": {
        "raw": "Example text"
      },
      "created_on": "2026-03-11T00:00:00Z",
      "deleted": true,
      "id": 1,
      "inline": {
        "path": "file.txt"
      },
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      },
      "parent": {
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "user": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        }
      },
      "pending": true,
      "resolution": {
        "user": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        }
      },
      "updated_on": "2026-03-11T00:00:00Z",
      "user": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      }
    },
    "pending": true,
    "resolution": {
      "created_on": "2026-03-11T00:00:00Z",
      "type": "commit_file",
      "user": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      }
    },
    "updated_on": "2026-03-11T00:00:00Z",
    "user": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    }
  },
  "deleted": true,
  "host": "bitbucket.org",
  "pull_request": 1,
  "repo": "repo-slug",
  "resolution": {
    "created_on": "2026-03-11T00:00:00Z",
    "type": "commit_file",
    "user": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    }
  },
  "workspace": "workspace-slug"
}
```

## Pull Request Tasks

Representative shapes for pull request task lists and task detail, create, edit, resolve, reopen, and delete payloads.

Commands:

```bash
bb pr task list 1 --repo workspace-slug/repo-slug --json '*'
bb pr task create 1 --repo workspace-slug/repo-slug --comment 15 --body 'Handle this thread' --json '*'
bb pr task resolve 3 --pr 1 --repo workspace-slug/repo-slug --json '*'
```

Representative shapes:

```json
{
  "host": "bitbucket.org",
  "pull_request": 1,
  "repo": "repo-slug",
  "state": "OPEN",
  "tasks": [
    {
      "comment": {
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "user": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        }
      },
      "content": {
        "html": "\u003chtml\u003e",
        "raw": "Example text"
      },
      "created_on": "2026-03-11T00:00:00Z",
      "creator": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      },
      "id": 1,
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      },
      "pending": true,
      "resolved_by": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      },
      "resolved_on": "\u003cresolved-on\u003e",
      "state": "OPEN",
      "updated_on": "2026-03-11T00:00:00Z"
    }
  ],
  "workspace": "workspace-slug"
}
```

```json
{
  "action": "\u003caction\u003e",
  "deleted": true,
  "host": "bitbucket.org",
  "pull_request": 1,
  "repo": "repo-slug",
  "task": {
    "comment": {
      "content": {
        "raw": "Example text"
      },
      "created_on": "2026-03-11T00:00:00Z",
      "deleted": true,
      "id": 1,
      "inline": {
        "path": "file.txt"
      },
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      },
      "parent": {
        "id": 1,
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        },
        "user": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        }
      },
      "pending": true,
      "resolution": {
        "user": {
          "account_id": "account-id",
          "display_name": "Example User",
          "nickname": "example-user"
        }
      },
      "updated_on": "2026-03-11T00:00:00Z",
      "user": {
        "account_id": "account-id",
        "display_name": "Example User",
        "nickname": "example-user"
      }
    },
    "content": {
      "html": "\u003chtml\u003e",
      "markup": "\u003cmarkup\u003e",
      "raw": "Example text"
    },
    "created_on": "2026-03-11T00:00:00Z",
    "creator": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "id": 1,
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      },
      "self": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    },
    "pending": true,
    "resolved_by": {
      "account_id": "account-id",
      "display_name": "Example User",
      "nickname": "example-user"
    },
    "resolved_on": "\u003cresolved-on\u003e",
    "state": "OPEN",
    "updated_on": "2026-03-11T00:00:00Z"
  },
  "workspace": "workspace-slug"
}
```

## Pull Request Status

Representative shape for the pull request status payload.

Command:

```bash
bb pr status --repo workspace-slug/repo-slug --json '*'
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
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspace": "workspace-slug"
}
```

## Pull Request Diff

Representative shape for the pull request diff payload.

Command:

```bash
bb pr diff 1 --repo workspace-slug/repo-slug --json '*'
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
  "warnings": [
    "\u003citem\u003e"
  ],
  "workspace": "workspace-slug"
}
```

## Issue List And View

Representative shape for issue list items and the issue view payload.

Commands:

```bash
bb issue list --repo workspace-slug/issues-repo-slug --json '*'
bb issue view 1 --repo workspace-slug/issues-repo-slug --json '*'
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
    },
    "self": {
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
bb search repos bb-cli --workspace workspace-slug --json '*'
bb search prs fixture --repo workspace-slug/repo-slug --json '*'
bb search issues broken --repo workspace-slug/issues-repo-slug --json '*'
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
  "parent": {
    "full_name": "workspace-slug/repo-slug",
    "name": "Example Name",
    "slug": "\u003cslug\u003e"
  },
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
      "participated_on": "\u003cparticipated-on\u003e",
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
    },
    "self": {
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
bb status --workspace workspace-slug --json '*'
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
