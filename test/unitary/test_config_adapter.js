
var assert = require('assert');
var config_adapter = require('../../src/lib/config_adapter');

describe('test config adapter', function() {
  it('should read the port from the yaml file', function() {
    yaml_config = {};
    yaml_config.port = 7070;
    var config = config_adapter(yaml_config); 
    assert.equal(config.port, 7070);
  });

  it('should default the port to 8080 if not provided', function() {
    yaml_config = {};
    var config = config_adapter(yaml_config); 
    assert.equal(config.port, 8080);
  });

  it('should get the ldap attributes', function() {
    yaml_config = {};
    yaml_config.ldap = {};
    yaml_config.ldap.url = 'http://ldap';
    yaml_config.ldap.user_search_base = 'ou=groups,dc=example,dc=com';
    yaml_config.ldap.user_search_filter = 'uid';
    yaml_config.ldap.user = 'admin';
    yaml_config.ldap.password = 'pass';

    var config = config_adapter(yaml_config); 

    assert.equal(config.ldap.url, 'http://ldap');
    assert.equal(config.ldap.user_search_base, 'ou=groups,dc=example,dc=com');
    assert.equal(config.ldap.user_search_filter, 'uid');
    assert.equal(config.ldap.user, 'admin');
    assert.equal(config.ldap.password, 'pass');
  });

  it('should get the session attributes', function() {
    yaml_config = {};
    yaml_config.session = {};
    yaml_config.session.domain = 'example.com';
    yaml_config.session.secret = 'secret';
    yaml_config.session.expiration = 3600;

    var config = config_adapter(yaml_config); 

    assert.equal(config.session_domain, 'example.com');
    assert.equal(config.session_secret, 'secret');
    assert.equal(config.session_max_age, 3600);
  });

  it('should get the log level', function() {
    yaml_config = {};
    yaml_config.logs_level = 'debug';

    var config = config_adapter(yaml_config); 
    assert.equal(config.logs_level, 'debug');
  });

  it('should get the notifier config', function() {
    yaml_config = {};
    yaml_config.notifier = 'notifier';

    var config = config_adapter(yaml_config); 
    
    assert.equal(config.notifier, 'notifier');
  });

  it('should get the access_control config', function() {
    yaml_config = {};
    yaml_config.access_control = 'access_control';

    var config = config_adapter(yaml_config); 
    
    assert.equal(config.access_control, 'access_control');
  });
});
