# Contributing

Anybody willing to contribute to the project either with code, documentation, security reviews or whatever, are very
welcome to create or review pull requests and take part in discussions in any of our public
[chat rooms](README.md#contact-options).

It's also possible to contribute financially in order to support the community.

Don't hesitate to come help us improve Authelia! See you soon!

## Bug Reports and Feature Requests

If you've found a **bug** or have a **feature request** then please create a
[bug report](https://www.authelia.com/l/bug) or [feature request](https://www.authelia.com/l/fr) respectively in this
repository (but search first in case a similar issue already exists).

## Code

If you would like to fix a bug or implement a feature, please fork the repository and create a Pull Request.
More information on getting set up locally can be found in the
[Development Contribution](https://www.authelia.com/contributing/development/introduction/) documentation, in addition
the [Contribution Guidelines](https://www.authelia.com/contributing/guidelines/introduction/) documentation includes
several contribution guidelines.

Before you start any Pull Request, it's recommended that you create an issue to discuss first if you have any doubts
about requirement or implementation. That way you can be sure that the maintainer(s) agree on what to change and how,
and you can hopefully get a quick merge afterwards. Also, let the maintainers know that you plan to work on a particular
issue so that no one else starts any duplicate work.

Pull Requests can only be merged once all status checks are green, which means `authelia-scripts --log-level debug ci`
passes, and coverage does not regress.

### Do not force push to your pull request branch

Please do not force push to your PR's branch after you have created your PR especially when a maintainer has either
performed a review or has indicated they are performing a review, as doing so makes it harder to review your commits
accurately. PRs will always be squashed by us when we merge your work. Commit as many times as you need in your
pull request branch.

A few exceptions exist to this rule and are as follows:

- Making adjustments to the commit message i.e. for the following reasons:
	- To comply with the [Commit Message] guidelines
- To rebase your changes off of master or another branch

[Commit Message]: https://www.authelia.com/contributing/guidelines/commit-message/

## Re-requesting a review

Please do not ping your reviewer(s) by mentioning them in a new comment.
Instead, use the re-request review functionality.
Read more about this in the [GitHub docs, Re-requesting a review](https://docs.github.com/en/free-pro-team@latest/github/collaborating-with-issues-and-pull-requests/incorporating-feedback-in-your-pull-request#re-requesting-a-review).

## Collaboration with maintainers

Sometimes the codebase can be a challenge to navigate, especially for a first-time contributor. We don't want you
spending an hour trying to work out something that would take us only a minute to explain.

If you'd like some help getting started we have several [contact options](README.md#contact-options) available.
