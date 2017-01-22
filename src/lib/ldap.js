
module.exports = {
  validate: validateCredentials,
  get_email: retrieve_email
}

var util = require('util');
var Promise = require('bluebird');
var exceptions = require('./exceptions');

function validateCredentials(ldap_client, username, password, users_dn) {
  var userDN = util.format("cn=%s,%s", username, users_dn);
  var bind_promised = Promise.promisify(ldap_client.bind, { context: ldap_client });
  return bind_promised(userDN, password)
  .error(function(err) {
    throw new exceptions.LdapBindError(err.message);
  });
}

function retrieve_email(ldap_client, username, users_dn) {
  var userDN = util.format("cn=%s,%s", username, users_dn);
  var search_promised = Promise.promisify(ldap_client.search, { context: ldap_client });
  var query = {};
  query.sizeLimit = 1;
  query.attributes = ['mail'];
  var base_dn = userDN;

  return new Promise(function(resolve, reject) {
    search_promised(base_dn, query)
    .then(function(res) {
      var doc;
      res.on('searchEntry', function(entry) {
        doc = entry.object;
      });
      res.on('error', function(err) {
        reject(new exceptions.LdapSearchError(err.message));
      });
      res.on('end', function(result) {
        resolve(doc);
      });
    })
    .catch(function(err) {
      reject(new exceptions.LdapSearchError(err.message));
    });
  });
}
