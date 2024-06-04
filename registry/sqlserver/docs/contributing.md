# Contributing

_First_: if you feel insecure about how to start contributing, feel free to ask us on our
[Discord channel](https://discordapp.com/invite/hasura) in the #contrib channel. You can also just go ahead with your contribution and we'll give you feedback. Don't worry - the worst that can happen is that you'll be politely asked to change something. We appreciate any contributions, and we don't want a wall of rules to stand in the way of that.

However, for those individuals who want a bit more guidance on the best way to contribute to the project, read on. This document will cover what we're looking for. By addressing the points below, the chances that we can quickly merge or address your contributions will increase.

## 1. Code of conduct

Please follow our [Code of conduct](./code-of-conduct.md) in the context of any contributions made to Hasura.

## 2. CLA

For all contributions, a CLA (Contributor License Agreement) needs to be signed
[here](https://cla-assistant.io/hasura/ndc-sqlserver) before (or after) the pull request has been submitted. A bot will prompt contributors to sign the CLA via a pull request comment, if necessary.

## 3. Ways of contributing

### Reporting an Issue

- Make sure you test against the latest released cloud version. It is possible that we may have already fixed the bug you're experiencing.
- Provide steps to reproduce the issue, including Database (e.g. SQL Server) version and Hasura DDN version.
- Please include logs, if relevant.
- Create a [issue](https://github.com/hasura/ndc-sqlserver/issues/new/choose).

### Working on an issue

- We use the [fork-and-branch git workflow](https://blog.scottlowe.org/2015/01/27/using-fork-branch-git-workflow/).
- Please make sure there is an issue associated with the work that you're doing.
- If you're working on an issue, please comment that you are doing so to prevent duplicate work by others also.
- Squash your commits and refer to the issue using `fix #<issue-no>` or `close #<issue-no>` in the commit message, at the end. For example: `resolve answers to everything (fix #42)` or `resolve answers to everything, fix #42`
- Rebase master with your branch before submitting a pull request.

## 6. Commit messages

- The first line should be a summary of the changes, not exceeding 50 characters, followed by an optional body which has more details about the changes. Refer to [this link](https://github.com/erlang/otp/wiki/writing-good-commit-messages) for more information on writing good commit messages.
- Use the imperative present tense: "add/fix/change", not "added/fixed/changed" nor "adds/fixes/changes".
- Don't capitalize the first letter of the summary line.
- Don't add a period/dot (.) at the end of the summary line.