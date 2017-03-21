
var objectPath = require('object-path');

module.exports = function(yaml_config) {
  return {
    port: objectPath.get(yaml_config, 'port', 8080),
    ldap_url: objectPath.get(yaml_config, 'ldap.url', 'ldap://127.0.0.1:389'),
    ldap_user_search_base: objectPath.get(yaml_config, 'ldap.user_search_base'),
    ldap_user_search_filter: objectPath.get(yaml_config, 'ldap.user_search_filter'),
    ldap_user: objectPath.get(yaml_config, 'ldap.user'),
    ldap_password: objectPath.get(yaml_config, 'ldap.password'),
    session_domain: objectPath.get(yaml_config, 'session.domain'),
    session_secret: objectPath.get(yaml_config, 'session.secret'),
    session_max_age: objectPath.get(yaml_config, 'session.expiration', 3600000), // in ms
    store_directory: objectPath.get(yaml_config, 'store_directory'),
    logs_level: objectPath.get(yaml_config, 'logs_level'),
    notifier: objectPath.get(yaml_config, 'notifier'),
  }
};

