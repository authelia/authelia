---
title: "authelia access-control check-policy"
description: "Reference for the authelia access-control check-policy command."
lead: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 905
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia access-control check-policy

Checks a request against the access control rules to determine what policy would be applied

### Synopsis


Checks a request against the access control rules to determine what policy would be applied.

Legend:

	#		The rule position in the configuration.
	*		The first fully matched rule.
	~		Potential match i.e. if the user was authenticated they may match this rule.
	hit     The criteria in this column is a match to the request.
	miss    The criteria in this column is not match to the request.
	may     The criteria in this column is potentially a match to the request.

Notes:

	A rule that potentially matches a request will cause a redirection to occur in order to perform one-factor
	authentication. This is so Authelia can adequately determine if the rule actually matches.


```
authelia access-control check-policy [flags]
```

### Examples

```
authelia access-control check-policy --config config.yml --url https://example.com
authelia access-control check-policy --config config.yml --url https://example.com --username john
authelia access-control check-policy --config config.yml --url https://example.com --groups admin,public
authelia access-control check-policy --config config.yml --url https://example.com --username john --method GET
authelia access-control check-policy --config config.yml --url https://example.com --username john --method GET --verbose
```

### Options

```
      --groups strings    the groups of the subject
  -h, --help              help for check-policy
      --ip string         the ip of the subject
      --method string     the HTTP method of the object (default "GET")
      --url string        the url of the object
      --username string   the username of the subject
      --verbose           enables verbose output
```

### Options inherited from parent commands

```
  -c, --config strings                        configuration files or directories to load, for more information run 'authelia -h authelia config' (default [configuration.yml])
      --config.experimental.filters strings   list of filters to apply to all configuration files, for more information run 'authelia -h authelia filters'
```

### SEE ALSO

* [authelia access-control](authelia_access-control.md)	 - Helpers for the access control system

