
module.exports = function(logger, acl_config) {
  return {
    builder: new AccessControlBuilder(logger, acl_config),
    matcher: new AccessControlMatcher(logger)
  };
}

var objectPath = require('object-path');

// *************** PER DOMAIN MATCHER ***************
function AccessControlMatcher(logger) {
  this.logger = logger;
}

AccessControlMatcher.prototype.is_domain_allowed = function(domain, allowed_domains) {
  // Allow all matcher
  if(allowed_domains.length == 1 && allowed_domains[0] == '*') return true;

  this.logger.debug('ACL: trying to match %s with %s', domain, 
                    JSON.stringify(allowed_domains));
  for(var i = 0; i < allowed_domains.length; ++i) {
    var allowed_domain = allowed_domains[i];
    if(allowed_domain.startsWith('*') && 
       domain.endsWith(allowed_domain.substr(1))) {
      return true;
    }
    else if(domain == allowed_domain) {
      return true;
    }
  }
  return false;
}


// *************** MATCHER BUILDER ***************
function AccessControlBuilder(logger, acl_config) {
  this.logger = logger;
  this.config = acl_config;
}

AccessControlBuilder.prototype.extract_per_group = function(groups) {
  var allowed_domains = [];
  var groups_policy = objectPath.get(this.config, 'groups');
  if(groups_policy) {
    for(var i=0; i<groups.length; ++i) {
      var group = groups[i];
      if(group in groups_policy) {
        allowed_domains = allowed_domains.concat(groups_policy[group]);
      }
    }
  }
  return allowed_domains;
}

AccessControlBuilder.prototype.extract_per_user = function(user) {
  var allowed_domains = [];
  var users_policy = objectPath.get(this.config, 'users');
  if(users_policy) {
    if(user in users_policy) {
      allowed_domains = allowed_domains.concat(users_policy[user]);
    }
  }
  return allowed_domains;
}

AccessControlBuilder.prototype.get_allowed_domains = function(user, groups) {
  var allowed_domains = [];
  var default_policy = objectPath.get(this.config, 'default');
  if(default_policy) {
    allowed_domains = allowed_domains.concat(default_policy);
  }

  allowed_domains = allowed_domains.concat(this.extract_per_group(groups));
  allowed_domains = allowed_domains.concat(this.extract_per_user(user));

  this.logger.debug('ACL: user \'%s\' is allowed access to %s', user, 
                    JSON.stringify(allowed_domains));
  return allowed_domains;
}

AccessControlBuilder.prototype.get_any_domain = function() {
  return ['*'];
}
