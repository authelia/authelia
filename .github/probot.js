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
