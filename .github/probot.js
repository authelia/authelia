// PR commentary for Authelia branch based contributions
on('pull_request.opened')
    .filter(
        context =>
            context.payload.pull_request.head.label.slice(0, 9) === 'authelia:'
    )
    .filter(
        context =>
            context.payload.pull_request.head.ref.slice(0, 11) !== 'dependabot/'
    )
    .comment(`# Docker Container
These changes are published for testing at the following location:
* \`docker pull authelia/authelia:{{ pull_request.head.ref }}\``)

// PR commentary for third party based contributions
on('pull_request.opened')
    .filter(
        context =>
            context.payload.pull_request.head.label.slice(0, 9) !== 'authelia:'
    )
    .comment(`# Docker Container
These changes once approved by a team member will be published for testing at the following location:
* \`docker pull authelia/authelia:PR{{ pull_request.number }}\``)