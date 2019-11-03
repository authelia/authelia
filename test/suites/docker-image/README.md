# Dockerhub suite

This suite is made to quickly test that the Docker image of Authelia runs properly when spawned.
It can also be used for you to test Authelia without building it since the latest image will be
pulled from Dockerhub.

## Components

This suite will spawn an highly-available setup with nginx, mongo, redis, OpenLDAP, etc...

## Tests

Check if the image runs and does not crash unexpectedly and do a simple authentication with 2FA.