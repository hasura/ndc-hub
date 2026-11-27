#!/usr/bin/env bash
set -euo pipefail
[[ "$GITHUB_REPOSITORY" == hasura/ndc-hub && "$GITHUB_REF" == refs/heads/main ]]
[[ "$PR_NUMBER" =~ ^[1-9][0-9]{0,8}$ && "$PR_HEAD_SHA" =~ ^[0-9a-f]{40}$ ]]
test "$(gh api "repos/$GITHUB_REPOSITORY/branches/main" --jq '.commit.sha')" = "$GITHUB_SHA"
gh api "repos/$GITHUB_REPOSITORY/pulls/$PR_NUMBER" | jq -e --arg sha "$PR_HEAD_SHA" '
  .state == "open" and (.draft | not) and .base.ref == "main" and
  .base.repo.full_name == "hasura/ndc-hub" and .head.sha == $sha' >/dev/null
