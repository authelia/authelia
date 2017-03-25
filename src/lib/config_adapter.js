
var objectPath = require('object-path');

module.exports = function(yaml_config) {
  return {
    port: objectPath.get(yaml_config, 'port', 8080),
    ldap: objectPath.get(yaml_config, 'ldap', 'ldap://127.0.0.1:389'),
    session_domain: objectPath.get(yaml_config, 'session.domain'),
    session_secret: objectPath.get(yaml_config, 'session.secret'),
    session_max_age: objectPath.get(yaml_config, 'session.expiration', 3600000), // in ms
    store_directory: objectPath.get(yaml_config, 'store_directory'),
    logs_level: objectPath.get(yaml_config, 'logs_level'),
    notifier: objectPath.get(yaml_config, 'notifier'),
    access_control: objectPath.get(yaml_config, 'access_control')
  }
};

