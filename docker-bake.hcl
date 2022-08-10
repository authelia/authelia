variable "IMAGE_TAG" {
	default = "bake"
}

group "default" {
	targets = ["base"]
}

target "base" {
	dockerfile = "Dockerfile"
	tags = ["docker.io/authelia/authelia:${IMAGE_TAG}", "ghcr.io/authelia/authelia:${IMAGE_TAG}"]
	platforms = ["linux/amd64", "linux/arm/v7", "linux/arm64"]
	labels = {
		"org.opencontainers.image.url" = "https://github.com/authelia/authelia/pkgs/container/authelia"
		"org.opencontainers.image.documentation" = "https://www.authelia.com"
		"org.opencontainers.image.vendor" = "Authelia"
		"org.opencontainers.image.licenses" = "Apache-2.0"
		"org.opencontainers.image.title" = "authelia"
		"org.opencontainers.image.description" = "Authelia is an open-source authentication and authorization server providing two-factor authentication and single sign-on (SSO) for your applications via a web portal."
	}
}
