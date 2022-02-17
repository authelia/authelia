---
layout: default
title: Integrate Authelia with Django
parent: Community
nav_order: 6
---

# Integrate Authelia with Django

Django, the Python web framework, can be configured to delegate authentication to external services 
using HTTP request headers. This is well documented on [Django documentation](https://docs.djangoproject.com/en/3.2/howto/auth-remote-user/)

Therefore, it is possible to integrate Django with Authelia following the documentation about 
[Proxy integration](https://www.authelia.com/docs/deployment/supported-proxies/#how-can-the-backend-be-aware-of-the-authenticated-users) 
and adding a few lines of code on your Django application.


## Basic integration

Django uses `REMOTE_USER` header by default. But WSGI servers transform the headers received from 
proxy servers adding `HTTP_` as prefix. So we need to add a custom middleware in order to use `HTTP_REMOTE_USER`.

This basic configuration enables authentication using Authelia. If the user does not exists on Django database,
it will be automatically created.


### Configuration

```python
# file: settings.py

MIDDLEWARE = [
    '...',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'your_app.auth.middleware.RemoteUserMiddleware',
    # or 'your_app.auth.middleware.PersistentRemoteUserMiddleware',
    '...',
]

AUTHENTICATION_BACKENDS = [
    'django.contrib.auth.backends.RemoteUserBackend',
]

# Logout from authelia after logout on the Django application
LOGOUT_REDIRECT_URL = 'https://auth.your_domain.com/logout'

```

### New authentication middleware

```python
# new file: your_app/auth/middleware.py
from django.contrib.auth.middleware import RemoteUserMiddleware, PersistentRemoteUserMiddleware


class HttpRemoteUserMiddleware(RemoteUserMiddleware):
    header = 'HTTP_REMOTE_USER'

    # uncomment the line below to disable authentication to users that not exists on Django database
    # create_unknown_user = False 


class PersistentHttpRemoteUserMiddleware(PersistentRemoteUserMiddleware):
    """
    The RemoteUserMiddleware authentication middleware assumes that the HTTP request header 
    REMOTE_USER is present with all authenticated requests.

    With PersistentRemoteUserMiddleware, it is possible to receive this header only on a few 
    pages (as login page) and maintain the authenticated session until explicit 
    logout by the user.
    """
    header = 'HTTP_REMOTE_USER' 

```

**Security Warning:**
The proxy server **must** set `Remote-User` header **every time** it hits the Django application. If you only
protect the login URL with Authelia and use the Persistent class, you have to set this header to `''` 
on the other locations.


## Advanced integration

While the basic integration only uses the HTTP header `Remote-User` set by Authelia, this advanced integration
uses also the HTTP headers `Remote-Name`, `Remote-Email` and `Remote-Groups`.

In this example, we create a new authentication backend on Django that will synchronize user data with Authelia 
backend, storing the name, the email and the groups of the user on the Django database.

### Configuration

```python
# file: settings.py

MIDDLEWARE = [
    '...',
    'django.contrib.auth.middleware.AuthenticationMiddleware',
    'your_app.auth.middleware.RemoteUserMiddleware',
    # or 'your_app.auth.middleware.PersistentRemoteUserMiddleware',
    '...',
]

AUTHENTICATION_BACKENDS = [
    'your_app.auth.backends.RemoteExtendedUserBackend',
]

# Logout from authelia after logout on the Django application
LOGOUT_REDIRECT_URL = 'https://auth.your_domain.com/logout'

```

### New authentication backend
```python
# new file: your_app/auth/backends.py
from django.conf import settings
from django.contrib.auth.models import Group
from django.contrib.auth.backends import RemoteUserBackend


class RemoteExtendedUserBackend(RemoteUserBackend):
    """
    This backend can be used in conjunction with the ``RemoteUserMiddleware``
    to handle authentication outside Django and update local user with external information
    (name, email and groups).

    Extends RemoteUserBackend (it creates the Django user if it does not exist,
    as explained here: https://github.com/django/django/blob/main/django/contrib/auth/backends.py#L167),
    updating the user with the information received from the remote headers.

    Django user is only added to groups that already exist on the database (no groups are created).
    A settings variable can be used to exclude some groups when updating the user.
    """

    excluded_groups = set()
    if hasattr(settings, 'REMOTE_AUTH_BACKEND_EXCLUDED_GROUPS'):
        excluded_groups = set(settings.REMOTE_AUTH_BACKEND_EXCLUDED_GROUPS)

    # Warning: possible security breach if reverse proxy does not set
    # these variables EVERY TIME it hits this Django application (and REMOTE_USER variable).
    # See https://docs.djangoproject.com/en/4.0/howto/auth-remote-user/#configuration
    header_name = 'HTTP_REMOTE_NAME'
    header_groups = 'HTTP_REMOTE_GROUPS'
    header_email = 'HTTP_REMOTE_EMAIL'

    def authenticate(self, request, remote_user):
        user = super().authenticate(request, remote_user)

        # original authenticate calls configure_user only
        # when user is created. We need to call this method every time
        # the user is authenticated in order to update its data.
        if user:
            self.configure_user(request, user)
        return user

    def configure_user(self, request, user):
        """
        Complete the user from extra request.META information.
        """
        if self.header_name in request.META:
            user.last_name = request.META[self.header_name]

        if self.header_email in request.META:
            user.email = request.META[self.header_email]

        if self.header_groups in request.META:
            self.update_groups(user, request.META[self.header_groups])

        if self.user_has_to_be_staff(user):
            user.is_staff = True

        user.save()
        return user

    def user_has_to_be_staff(self, user):
        return True

    def update_groups(self, user, remote_groups):
        """
        Synchronizes groups the user belongs to with remote information.

        Groups (existing django groups or remote groups) on excluded_groups are completely ignored.
        No group will be created on the django database.

        Disclaimer: this method is strongly inspired by the LDAPBackend from django-auth-ldap.
        """
        current_group_names = frozenset(
            user.groups.values_list("name", flat=True).iterator()
        )
        preserved_group_names = current_group_names.intersection(self.excluded_groups)
        current_group_names = current_group_names - self.excluded_groups

        target_group_names = frozenset(
            [x for x in map(self.clean_groupname, remote_groups.split(',')) if x is not None]
        )
        target_group_names = target_group_names - self.excluded_groups

        if target_group_names != current_group_names:
            target_group_names = target_group_names.union(preserved_group_names)
            existing_groups = list(
                Group.objects.filter(name__in=target_group_names).iterator()
            )
            user.groups.set(existing_groups)
        return

    def clean_groupname(self, groupname):
        """
        Perform any cleaning on the "groupname" prior to using it.
        Return the cleaned groupname.
        """
        return groupname

```
