
module.exports = {
  validate: validate_credentials,
  get_email: retrieve_email,
  update_password: update_password
}

var util = require('util');
var Promise = require('bluebird');
var exceptions = require('./exceptions');
var Dovehash = require('dovehash');

function validate_credentials(ldap_client, username, password, user_base, user_filter) {
  // if not provided, default to cn
  if(!user_filter) user_filter = 'cn';

  var userDN = util.format("%s=%s,%s", user_filter, username, user_base);
  console.log(userDN);
  var bind_promised = Promise.promisify(ldap_client.bind, { context: ldap_client });
  return bind_promised(userDN, password)
  .error(function(err) {
    console.error(err);
    throw new exceptions.LdapBindError(err.message);
  });
}

function retrieve_email(ldap_client, username, user_base, user_filter) {
  // if not provided, default to cn
  if(!user_filter) user_filter = 'cn';

  var userDN = util.format("%s=%s,%s", user_filter, username, user_base);
  console.log(userDN);
  var search_promised = Promise.promisify(ldap_client.search, { context: ldap_client });
  var query = {};
  query.sizeLimit = 1;
  query.attributes = ['mail'];

  return new Promise(function(resolve, reject) {
    search_promised(userDN, query)
    .then(function(res) {
      var doc;
      res.on('searchEntry', function(entry) {
        doc = entry.object;
      });
      res.on('error', function(err) {
        reject(new exceptions.LdapSearchError(err));
      });
      res.on('end', function(result) {
        resolve(doc);
      });
    })
    .catch(function(err) {
      reject(new exceptions.LdapSearchError(err));
    });
  });
}

function update_password(ldap_client, ldap, username, new_password, config) {
  var user_filter = config.ldap_user_search_filter;
  // if not provided, default to cn
  if(!user_filter) user_filter = 'cn';

  var userDN = util.format("%s=%s,%s", user_filter, username, 
                           config.ldap_user_search_base);
  var encoded_password = Dovehash.encode('SSHA', new_password);
  var change = new ldap.Change({
    operation: 'replace',
    modification: {
      userPassword: encoded_password
    }
  });
  
  var modify_promised = Promise.promisify(ldap_client.modify, { context: ldap_client });
  var bind_promised = Promise.promisify(ldap_client.bind, { context: ldap_client });

  return bind_promised(config.ldap_user, config.ldap_password)
  .then(function() {
    return modify_promised(userDN, change);
  });
}
