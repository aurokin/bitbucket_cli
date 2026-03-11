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
  "host": "bitbucket.org",
  "workspace": "workspace-slug",
  "repo": "repo-slug",
  "name": "Example Name",
  "full_name": "workspace-slug/repo-slug",
  "description": "Example text",
  "private": true,
  "project_key": "BBCLI",
  "project_name": "Bitbucket CLI",
  "main_branch": "main",
  "html_url": "https://bitbucket.org/workspace-slug/repo-slug",
  "https_clone": "https://bitbucket.org/workspace-slug/repo-slug.git",
  "ssh_clone": "git@bitbucket.org:workspace-slug/repo-slug.git",
  "remote": "origin",
  "local_clone_url": "git@bitbucket.org:workspace-slug/repo-slug.git",
  "root": "/path/to/repo"
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
  "host": "bitbucket.org",
  "workspace": "workspace-slug",
  "repo": "repo-slug",
  "name": "Example Name",
  "directory": "/path/to/repo",
  "clone_url": "https://bitbucket.org/workspace-slug/repo-slug.git"
}
```

```json
{
  "host": "bitbucket.org",
  "workspace": "workspace-slug",
  "repo": "repo-slug",
  "name": "Example Name",
  "deleted": true
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
  "id": 1,
  "title": "Example title",
  "description": "Example text",
  "state": "OPEN",
  "author": {
    "display_name": "Example User",
    "nickname": "example-user",
    "account_id": "account-id"
  },
  "reviewers": [
    {
      "display_name": "Example User",
      "nickname": "example-user",
      "account_id": "account-id"
    }
  ],
  "participants": [
    {
      "user": {
        "display_name": "Example User",
        "nickname": "example-user",
        "account_id": "account-id"
      },
      "role": "REVIEWER",
      "approved": true,
      "state": "OPEN"
    }
  ],
  "source": {
    "branch": {
      "name": "main",
      "merge_strategies": [
        "\u003citem\u003e"
      ],
      "default_merge_strategy": "merge_commit"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "name": "Example Name",
      "full_name": "workspace-slug/repo-slug"
    }
  },
  "destination": {
    "branch": {
      "name": "main",
      "merge_strategies": [
        "\u003citem\u003e"
      ],
      "default_merge_strategy": "merge_commit"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "name": "Example Name",
      "full_name": "workspace-slug/repo-slug"
    }
  },
  "close_source_branch": true,
  "queued": true,
  "merge_commit": {
    "hash": "abc123def456"
  },
  "draft": true,
  "comment_count": 2,
  "task_count": 2,
  "updated_on": "2026-03-11T00:00:00Z",
  "created_on": "2026-03-11T00:00:00Z",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  }
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
  "host": "bitbucket.org",
  "workspace": "workspace-slug",
  "repo": "repo-slug",
  "current_user": {
    "account_id": "account-id",
    "display_name": "Example User",
    "username": "user@example.com",
    "uuid": "{uuid}"
  },
  "current_branch_name": "\u003ccurrent-branch-name\u003e",
  "current_branch": {
    "id": 1,
    "title": "Example title",
    "description": "Example text",
    "state": "OPEN",
    "author": {
      "display_name": "Example User",
      "nickname": "example-user",
      "account_id": "account-id"
    },
    "reviewers": [
      {
        "display_name": "Example User",
        "nickname": "example-user",
        "account_id": "account-id"
      }
    ],
    "participants": [
      {
        "user": {
          "display_name": "Example User",
          "nickname": "example-user",
          "account_id": "account-id"
        },
        "role": "REVIEWER",
        "approved": true,
        "state": "OPEN"
      }
    ],
    "source": {
      "branch": {
        "name": "main",
        "merge_strategies": [
          "\u003citem\u003e"
        ],
        "default_merge_strategy": "merge_commit"
      },
      "commit": {
        "hash": "abc123def456"
      },
      "repository": {
        "name": "Example Name",
        "full_name": "workspace-slug/repo-slug"
      }
    },
    "destination": {
      "branch": {
        "name": "main",
        "merge_strategies": [
          "\u003citem\u003e"
        ],
        "default_merge_strategy": "merge_commit"
      },
      "commit": {
        "hash": "abc123def456"
      },
      "repository": {
        "name": "Example Name",
        "full_name": "workspace-slug/repo-slug"
      }
    },
    "close_source_branch": true,
    "queued": true,
    "merge_commit": {
      "hash": "abc123def456"
    },
    "draft": true,
    "comment_count": 2,
    "task_count": 2,
    "updated_on": "2026-03-11T00:00:00Z",
    "created_on": "2026-03-11T00:00:00Z",
    "links": {
      "html": {
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    }
  },
  "created": [
    {
      "id": 1,
      "title": "Example title",
      "description": "Example text",
      "state": "OPEN",
      "author": {
        "display_name": "Example User",
        "nickname": "example-user",
        "account_id": "account-id"
      },
      "reviewers": [
        {
          "display_name": "Example User",
          "nickname": "example-user",
          "account_id": "account-id"
        }
      ],
      "participants": [
        {
          "user": {
            "display_name": "Example User",
            "nickname": "example-user",
            "account_id": "account-id"
          },
          "role": "REVIEWER",
          "approved": true,
          "state": "OPEN"
        }
      ],
      "source": {
        "branch": {
          "name": "main",
          "merge_strategies": [
            "\u003citem\u003e"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "commit": {
          "hash": "abc123def456"
        },
        "repository": {
          "name": "Example Name",
          "full_name": "workspace-slug/repo-slug"
        }
      },
      "destination": {
        "branch": {
          "name": "main",
          "merge_strategies": [
            "\u003citem\u003e"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "commit": {
          "hash": "abc123def456"
        },
        "repository": {
          "name": "Example Name",
          "full_name": "workspace-slug/repo-slug"
        }
      },
      "close_source_branch": true,
      "queued": true,
      "merge_commit": {
        "hash": "abc123def456"
      },
      "draft": true,
      "comment_count": 2,
      "task_count": 2,
      "updated_on": "2026-03-11T00:00:00Z",
      "created_on": "2026-03-11T00:00:00Z",
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      }
    }
  ],
  "review_requested": [
    {
      "id": 1,
      "title": "Example title",
      "description": "Example text",
      "state": "OPEN",
      "author": {
        "display_name": "Example User",
        "nickname": "example-user",
        "account_id": "account-id"
      },
      "reviewers": [
        {
          "display_name": "Example User",
          "nickname": "example-user",
          "account_id": "account-id"
        }
      ],
      "participants": [
        {
          "user": {
            "display_name": "Example User",
            "nickname": "example-user",
            "account_id": "account-id"
          },
          "role": "REVIEWER",
          "approved": true,
          "state": "OPEN"
        }
      ],
      "source": {
        "branch": {
          "name": "main",
          "merge_strategies": [
            "\u003citem\u003e"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "commit": {
          "hash": "abc123def456"
        },
        "repository": {
          "name": "Example Name",
          "full_name": "workspace-slug/repo-slug"
        }
      },
      "destination": {
        "branch": {
          "name": "main",
          "merge_strategies": [
            "\u003citem\u003e"
          ],
          "default_merge_strategy": "merge_commit"
        },
        "commit": {
          "hash": "abc123def456"
        },
        "repository": {
          "name": "Example Name",
          "full_name": "workspace-slug/repo-slug"
        }
      },
      "close_source_branch": true,
      "queued": true,
      "merge_commit": {
        "hash": "abc123def456"
      },
      "draft": true,
      "comment_count": 2,
      "task_count": 2,
      "updated_on": "2026-03-11T00:00:00Z",
      "created_on": "2026-03-11T00:00:00Z",
      "links": {
        "html": {
          "href": "https://bitbucket.org/workspace-slug/repo-slug"
        }
      }
    }
  ]
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
  "workspace": "workspace-slug",
  "repo": "repo-slug",
  "id": 1,
  "title": "Example title",
  "patch": "diff --git a/file.txt b/file.txt\n...",
  "stats": [
    {
      "status": "modified",
      "old": {
        "path": "file.txt",
        "escaped_path": "file.txt",
        "type": "commit_file"
      },
      "new": {
        "path": "file.txt",
        "escaped_path": "file.txt",
        "type": "commit_file"
      },
      "lines_added": 1,
      "lines_removed": 1
    }
  ]
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
  "id": 1,
  "title": "Example title",
  "state": "OPEN",
  "kind": "\u003ckind\u003e",
  "priority": "\u003cpriority\u003e",
  "content": {
    "raw": "Example text"
  },
  "reporter": {
    "display_name": "Example User",
    "account_id": "account-id",
    "nickname": "example-user"
  },
  "assignee": {
    "display_name": "Example User",
    "account_id": "account-id",
    "nickname": "example-user"
  },
  "created_on": "2026-03-11T00:00:00Z",
  "updated_on": "2026-03-11T00:00:00Z",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  }
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
  "name": "Example Name",
  "slug": "\u003cslug\u003e",
  "full_name": "workspace-slug/repo-slug",
  "description": "Example text",
  "is_private": true,
  "updated_on": "2026-03-11T00:00:00Z",
  "project": {
    "key": "BBCLI",
    "name": "BBCLI"
  },
  "mainbranch": {
    "name": "Example Name"
  },
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    },
    "clone": [
      {
        "name": "Example Name",
        "href": "https://bitbucket.org/workspace-slug/repo-slug"
      }
    ]
  }
}
```

```json
{
  "id": 1,
  "title": "Example title",
  "description": "Example text",
  "state": "OPEN",
  "author": {
    "display_name": "Example User",
    "nickname": "example-user",
    "account_id": "account-id"
  },
  "reviewers": [
    {
      "display_name": "Example User",
      "nickname": "example-user",
      "account_id": "account-id"
    }
  ],
  "participants": [
    {
      "user": {
        "display_name": "Example User",
        "nickname": "example-user",
        "account_id": "account-id"
      },
      "role": "REVIEWER",
      "approved": true,
      "state": "OPEN"
    }
  ],
  "source": {
    "branch": {
      "name": "main",
      "merge_strategies": [
        "\u003citem\u003e"
      ],
      "default_merge_strategy": "merge_commit"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "name": "Example Name",
      "full_name": "workspace-slug/repo-slug"
    }
  },
  "destination": {
    "branch": {
      "name": "main",
      "merge_strategies": [
        "\u003citem\u003e"
      ],
      "default_merge_strategy": "merge_commit"
    },
    "commit": {
      "hash": "abc123def456"
    },
    "repository": {
      "name": "Example Name",
      "full_name": "workspace-slug/repo-slug"
    }
  },
  "close_source_branch": true,
  "queued": true,
  "merge_commit": {
    "hash": "abc123def456"
  },
  "draft": true,
  "comment_count": 2,
  "task_count": 2,
  "updated_on": "2026-03-11T00:00:00Z",
  "created_on": "2026-03-11T00:00:00Z",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  }
}
```

```json
{
  "id": 1,
  "title": "Example title",
  "state": "OPEN",
  "kind": "\u003ckind\u003e",
  "priority": "\u003cpriority\u003e",
  "content": {
    "raw": "Example text"
  },
  "reporter": {
    "display_name": "Example User",
    "account_id": "account-id",
    "nickname": "example-user"
  },
  "assignee": {
    "display_name": "Example User",
    "account_id": "account-id",
    "nickname": "example-user"
  },
  "created_on": "2026-03-11T00:00:00Z",
  "updated_on": "2026-03-11T00:00:00Z",
  "links": {
    "html": {
      "href": "https://bitbucket.org/workspace-slug/repo-slug"
    }
  }
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
  "user": "Example User",
  "workspaces": [
    "\u003citem\u003e"
  ],
  "repositories_scanned": 4,
  "repo_limit_per_workspace": 20,
  "item_limit_per_section": 20,
  "authored_prs_total": 3,
  "review_requested_prs_total": 3,
  "your_issues_total": 3,
  "repositories_without_issue_tracker": 1,
  "workspaces_at_repo_limit": [
    "\u003citem\u003e"
  ],
  "warnings": [
    "\u003citem\u003e"
  ],
  "authored_prs": [
    {
      "workspace": "workspace-slug",
      "repo": "repo-slug",
      "pull_request": {
        "id": 1,
        "title": "Example title",
        "description": "Example text",
        "state": "OPEN",
        "author": {
          "display_name": "Example User",
          "nickname": "example-user",
          "account_id": "account-id"
        },
        "reviewers": [
          {
            "display_name": "Example User",
            "nickname": "example-user",
            "account_id": "account-id"
          }
        ],
        "participants": [
          {
            "user": {
              "display_name": "Example User",
              "nickname": "example-user",
              "account_id": "account-id"
            },
            "role": "REVIEWER",
            "approved": true,
            "state": "OPEN"
          }
        ],
        "source": {
          "branch": {
            "name": "main",
            "merge_strategies": [
              "\u003citem\u003e"
            ],
            "default_merge_strategy": "merge_commit"
          },
          "commit": {
            "hash": "abc123def456"
          },
          "repository": {
            "name": "Example Name",
            "full_name": "workspace-slug/repo-slug"
          }
        },
        "destination": {
          "branch": {
            "name": "main",
            "merge_strategies": [
              "\u003citem\u003e"
            ],
            "default_merge_strategy": "merge_commit"
          },
          "commit": {
            "hash": "abc123def456"
          },
          "repository": {
            "name": "Example Name",
            "full_name": "workspace-slug/repo-slug"
          }
        },
        "close_source_branch": true,
        "queued": true,
        "merge_commit": {
          "hash": "abc123def456"
        },
        "draft": true,
        "comment_count": 2,
        "task_count": 2,
        "updated_on": "2026-03-11T00:00:00Z",
        "created_on": "2026-03-11T00:00:00Z",
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        }
      }
    }
  ],
  "review_requested_prs": [
    {
      "workspace": "workspace-slug",
      "repo": "repo-slug",
      "pull_request": {
        "id": 1,
        "title": "Example title",
        "description": "Example text",
        "state": "OPEN",
        "author": {
          "display_name": "Example User",
          "nickname": "example-user",
          "account_id": "account-id"
        },
        "reviewers": [
          {
            "display_name": "Example User",
            "nickname": "example-user",
            "account_id": "account-id"
          }
        ],
        "participants": [
          {
            "user": {
              "display_name": "Example User",
              "nickname": "example-user",
              "account_id": "account-id"
            },
            "role": "REVIEWER",
            "approved": true,
            "state": "OPEN"
          }
        ],
        "source": {
          "branch": {
            "name": "main",
            "merge_strategies": [
              "\u003citem\u003e"
            ],
            "default_merge_strategy": "merge_commit"
          },
          "commit": {
            "hash": "abc123def456"
          },
          "repository": {
            "name": "Example Name",
            "full_name": "workspace-slug/repo-slug"
          }
        },
        "destination": {
          "branch": {
            "name": "main",
            "merge_strategies": [
              "\u003citem\u003e"
            ],
            "default_merge_strategy": "merge_commit"
          },
          "commit": {
            "hash": "abc123def456"
          },
          "repository": {
            "name": "Example Name",
            "full_name": "workspace-slug/repo-slug"
          }
        },
        "close_source_branch": true,
        "queued": true,
        "merge_commit": {
          "hash": "abc123def456"
        },
        "draft": true,
        "comment_count": 2,
        "task_count": 2,
        "updated_on": "2026-03-11T00:00:00Z",
        "created_on": "2026-03-11T00:00:00Z",
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        }
      }
    }
  ],
  "your_issues": [
    {
      "workspace": "workspace-slug",
      "repo": "repo-slug",
      "issue": {
        "id": 1,
        "title": "Example title",
        "state": "OPEN",
        "kind": "\u003ckind\u003e",
        "priority": "\u003cpriority\u003e",
        "content": {
          "raw": "Example text"
        },
        "reporter": {
          "display_name": "Example User",
          "account_id": "account-id",
          "nickname": "example-user"
        },
        "assignee": {
          "display_name": "Example User",
          "account_id": "account-id",
          "nickname": "example-user"
        },
        "created_on": "2026-03-11T00:00:00Z",
        "updated_on": "2026-03-11T00:00:00Z",
        "links": {
          "html": {
            "href": "https://bitbucket.org/workspace-slug/repo-slug"
          }
        }
      }
    }
  ]
}
```
