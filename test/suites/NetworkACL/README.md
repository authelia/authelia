# Basic suite

This suite has been created to test Authelia with basic feature in a non highly-available setup.
Authelia basically use an in-memory cache to store user sessions and persist data on disk instead
of using a remote database. Also, the user accounts are stored in file-based database.

## Components

Authelia, nginx, fake webmail for registering devices.

## Tests

Broad range of tests.