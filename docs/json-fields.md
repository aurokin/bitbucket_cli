# JSON Field Index

Generated from the current payload structs and Bitbucket response models.

Use this file to discover top-level field names for `--json` selection.

| Command | Top-level fields | Example |
|---|---|---|
| `bb repo view` | `description`, `full_name`, `host`, `html_url`, `https_clone`, `local_clone_url`, `main_branch`, `name`, `private`, `project_key`, `project_name`, `remote`, `repo`, `root`, `ssh_clone`, `warnings`, `workspace` | `bb repo view --json description,full_name,host` |
| `bb repo clone` | `clone_url`, `directory`, `host`, `name`, `repo`, `workspace` | `bb repo clone --json clone_url,directory,host` |
| `bb repo delete` | `deleted`, `host`, `name`, `repo`, `workspace` | `bb repo delete --json deleted,host,name` |
| `bb browse` | `commit`, `host`, `issue`, `line`, `opened`, `path`, `pr`, `ref`, `repo`, `type`, `url`, `warnings`, `workspace` | `bb browse --json commit,host,issue` |
| `bb resolve` | `canonical_url`, `comment`, `commit`, `host`, `issue`, `line`, `path`, `pr`, `ref`, `repo`, `type`, `url`, `workspace` | `bb resolve --json canonical_url,comment,commit` |
| `bb pipeline list` | `build_number`, `completed_on`, `created_on`, `creator`, `links`, `state`, `target`, `uuid` | `bb pipeline list --json build_number,completed_on,created_on` |
| `bb pipeline log` | `host`, `log`, `pipeline`, `repo`, `step`, `workspace` | `bb pipeline log --json host,log,pipeline` |
| `bb pipeline stop` | `host`, `pipeline`, `repo`, `stopped`, `workspace` | `bb pipeline stop --json host,pipeline,repo` |
| `bb pipeline view` | `host`, `pipeline`, `repo`, `steps`, `workspace` | `bb pipeline view --json host,pipeline,repo` |
| `bb pr list` | `author`, `close_source_branch`, `comment_count`, `created_on`, `description`, `destination`, `draft`, `id`, `links`, `merge_commit`, `participants`, `queued`, `reviewers`, `source`, `state`, `task_count`, `title`, `updated_on` | `bb pr list --json author,close_source_branch,comment_count` |
| `bb pr view` | `author`, `close_source_branch`, `comment_count`, `created_on`, `description`, `destination`, `draft`, `id`, `links`, `merge_commit`, `participants`, `queued`, `reviewers`, `source`, `state`, `task_count`, `title`, `updated_on` | `bb pr view --json author,close_source_branch,comment_count` |
| `bb pr status` | `created`, `current_branch`, `current_branch_error`, `current_branch_name`, `current_user`, `host`, `repo`, `review_requested`, `warnings`, `workspace` | `bb pr status --json created,current_branch,current_branch_error` |
| `bb pr diff` | `host`, `id`, `patch`, `repo`, `stats`, `title`, `workspace` | `bb pr diff --json host,id,patch` |
| `bb issue list` | `assignee`, `content`, `created_on`, `id`, `kind`, `links`, `priority`, `reporter`, `state`, `title`, `updated_on` | `bb issue list --json assignee,content,created_on` |
| `bb issue view` | `assignee`, `content`, `created_on`, `id`, `kind`, `links`, `priority`, `reporter`, `state`, `title`, `updated_on` | `bb issue view --json assignee,content,created_on` |
| `bb search repos` | `description`, `full_name`, `is_private`, `links`, `mainbranch`, `name`, `project`, `slug`, `updated_on` | `bb search repos --json description,full_name,is_private` |
| `bb search prs` | `author`, `close_source_branch`, `comment_count`, `created_on`, `description`, `destination`, `draft`, `id`, `links`, `merge_commit`, `participants`, `queued`, `reviewers`, `source`, `state`, `task_count`, `title`, `updated_on` | `bb search prs --json author,close_source_branch,comment_count` |
| `bb search issues` | `assignee`, `content`, `created_on`, `id`, `kind`, `links`, `priority`, `reporter`, `state`, `title`, `updated_on` | `bb search issues --json assignee,content,created_on` |
| `bb status` | `authored_prs`, `authored_prs_total`, `item_limit_per_section`, `repo_limit_per_workspace`, `repositories_scanned`, `repositories_without_issue_tracker`, `review_requested_prs`, `review_requested_prs_total`, `user`, `warnings`, `workspaces`, `workspaces_at_repo_limit`, `your_issues`, `your_issues_total` | `bb status --json authored_prs,authored_prs_total,item_limit_per_section` |
| `bb auth status` | `default_host`, `hosts` | `bb auth status --json default_host,hosts` |
| `bb config list` | `key`, `source`, `value` | `bb config list --json key,source,value` |
| `bb alias list` | `expansion`, `name` | `bb alias list --json expansion,name` |
| `bb extension list` | `executable`, `name` | `bb extension list --json executable,name` |
