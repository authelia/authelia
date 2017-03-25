
module.exports = Ldap;

var util = require('util');
var Promise = require('bluebird');
var exceptions = require('./exceptions');
var Dovehash = require('dovehash');

function Ldap(deps, ldap_config) {
  this.ldap_config = ldap_config;

  this.ldapjs = deps.ldapjs;
  this.logger = deps.winston;

  this.connect();
}

Ldap.prototype.connect = function() {
  var ldap_client = this.ldapjs.createClient({
    url: this.ldap_config.url,
    reconnect: true
  });
  
  ldap_client.on('error', function(err) {
    console.error('LDAP Error:', err.message)
  });

  this.ldap_client = Promise.promisifyAll(ldap_client);
}

Ldap.prototype._build_user_dn = function(username) {
  var user_name_attr = this.ldap_config.user_name_attribute;
  // if not provided, default to cn
  if(!user_name_attr) user_name_attr = 'cn';

  var additional_user_dn = this.ldap_config.additional_user_dn;
  var base_dn = this.ldap_config.base_dn;

  var user_dn = util.format("%s=%s", user_name_attr, username);
  if(additional_user_dn) user_dn += util.format(",%s", additional_user_dn);
  user_dn += util.format(',%s', base_dn);
  return user_dn;
}

Ldap.prototype.bind = function(username, password) {
  var user_dn = this._build_user_dn(username);

  this.logger.debug('LDAP: Bind user %s', user_dn);
  return this.ldap_client.bindAsync(user_dn, password)
  .error(function(err) {
    throw new exceptions.LdapBindError(err.message);
  });
}

Ldap.prototype._search_in_ldap = function(base, query) {
  var that = this;
  this.logger.debug('LDAP: Search for %s in %s', JSON.stringify(query), base);
  return new Promise(function(resolve, reject) {
    that.ldap_client.searchAsync(base, query)
    .then(function(res) {
      var doc = [];
      res.on('searchEntry', function(entry) {
        doc.push(entry.object);
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

Ldap.prototype.get_groups = function(username) {
  var user_dn = this._build_user_dn(username);

  var group_name_attr = this.ldap_config.group_name_attribute;
  if(!group_name_attr) group_name_attr = 'cn';

  var additional_group_dn = this.ldap_config.additional_group_dn;
  var base_dn = this.ldap_config.base_dn;

  var group_dn = base_dn;
  if(additional_group_dn)
    group_dn = util.format('%s,', additional_group_dn) + group_dn;

  var query = {};
  query.scope = 'sub';
  query.attributes = [group_name_attr];
  query.filter = 'member=' + user_dn ;

  var that = this;
  this.logger.debug('LDAP: get groups of user %s', username);
  return this._search_in_ldap(group_dn, query)
  .then(function(docs) {
    var groups = [];
    for(var i = 0; i<docs.length; ++i) {
      groups.push(docs[i].cn);
    }
    that.logger.debug('LDAP: got groups %s', groups);
    return Promise.resolve(groups);
  });
}

Ldap.prototype.get_emails = function(username) {
  var that = this;
  var user_dn = this._build_user_dn(username);

  var query = {};
  query.scope = 'base';
  query.sizeLimit = 1;
  query.attributes = ['mail'];

  this.logger.debug('LDAP: get emails of user %s', username);
  return this._search_in_ldap(user_dn, query)
  .then(function(docs) {
    var emails = [];
    for(var i = 0; i<docs.length; ++i) {
      if(typeof docs[i].mail === 'string')
        emails.push(docs[i].mail);
      else {
        emails.concat(docs[i].mail);
      }
    }
    that.logger.debug('LDAP: got emails %s', emails);
    return Promise.resolve(emails);
  });
}

Ldap.prototype.update_password = function(username, new_password) {
  var user_dn = this._build_user_dn(username);

  var encoded_password = Dovehash.encode('SSHA', new_password);
  var change = new this.ldapjs.Change({
    operation: 'replace',
    modification: {
      userPassword: encoded_password
    }
  });
  
  var that = this;
  this.logger.debug('LDAP: update password of user %s', username);

  this.logger.debug('LDAP: bind admin');
  return this.ldap_client.bindAsync(this.ldap_config.user, this.ldap_config.password)
  .then(function() {
    that.logger.debug('LDAP: modify password');
    return that.ldap_client.modifyAsync(user_dn, change);
  });
}
