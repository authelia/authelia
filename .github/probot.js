// PR commentary for Authelia branch based contributions
on('pull_request.opened')
    .filter(
        context =>
            context.payload.pull_request.head.label.slice(0, 9) === 'authelia:'
    )
    .filter(
        context =>
            context.payload.pull_request.head.ref.slice(0, 9) !== 'renovate/'
    )
    .comment(`## Artifacts
These changes are published for testing on Buildkite and DockerHub.

### Docker Container
* \`docker pull authelia/authelia:{{ pull_request.head.ref }}\``)

// PR commentary for third party based contributions
on('pull_request.opened')
    .filter(
        context =>
            context.payload.pull_request.head.label.slice(0, 9) !== 'authelia:'
    )
    .comment(`Thanks for choosing to contribute @{{ pull_request.user.login }}. We lint all PR's with golangci-lint and eslint, I may add a review to your PR with some suggestions.
    
You are free to apply the changes if you're comfortable, alternatively you are welcome to ask a team member for advice.

## Artifacts
These changes once approved by a team member will be published for testing on Buildkite and DockerHub.

### Docker Container
* \`docker pull authelia/authelia:PR{{ pull_request.number }}\``)