// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

const excludedBranchPrefixes =
    /^(docs|all-contributors\/)/;

// PR commentary for Authelia branch based contributions
on("pull_request.opened")
    .filter((context) =>
        context.payload.pull_request.head.label.startsWith("authelia:"),
    )
    .filter((context) => {
        return !excludedBranchPrefixes.test(context.payload.pull_request.head.ref);
    })
    .filter((context) => !context.payload.pull_request.title.startsWith("docs"))
    .comment(`## Artifacts
These changes are published for testing on Buildkite, DockerHub and GitHub Container Registry.

### Docker Container
* \`docker pull authelia/authelia:{{ pull_request.head.ref }}\`
* \`docker pull ghcr.io/authelia/authelia:{{ pull_request.head.ref }}\``);

// PR commentary for third party based contributions
on("pull_request.opened").filter((context) =>
    !context.payload.pull_request.head.label.startsWith("authelia:"),
)
    .comment(`Thanks for choosing to contribute @{{ pull_request.user.login }}. We lint all PR's with golangci-lint and eslint, I may add a review to your PR with some suggestions.

You are free to apply the changes if you're comfortable, alternatively you are welcome to ask a team member for advice.

## Artifacts
These changes once approved by a team member will be published for testing on Buildkite, DockerHub and GitHub Container Registry.

### Docker Container
* \`docker pull authelia/authelia:PR{{ pull_request.number }}\`
* \`docker pull ghcr.io/authelia/authelia:PR{{ pull_request.number }}\``);

// Maintainer notification for a contributor who has not committed to the repository before, so they
// can be credited once their work merges.
on("pull_request.opened")
    .filter((context) =>
        ["FIRST_TIME_CONTRIBUTOR", "FIRST_TIMER"].includes(
            context.payload.pull_request.author_association,
        ),
    )
    .comment(`@authelia/review-general this is the first contribution to Authelia from @{{ pull_request.user.login }}.

Once it merges please credit them with the [all-contributors bot](https://allcontributors.org/en/bot/usage), picking the types from the [emoji key](https://allcontributors.org/en/reference/emoji-key/).`);
