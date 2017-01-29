
module.exports = UserDataStore;

var Promise = require('bluebird');
var path = require('path');

function UserDataStore(DataStore, options) {
  this._u2f_meta_collection = create_collection('u2f_meta', options, DataStore);
  this._identity_check_tokens_collection = 
    create_collection('identity_check_tokens', options, DataStore);
  this._authentication_traces_collection = 
    create_collection('authentication_traces', options, DataStore);
  this._totp_secret_collection = 
    create_collection('totp_secrets', options, DataStore);
}

function create_collection(name, options, DataStore) {
  var datastore_options = {};

  if(options.directory) 
    datastore_options.filename = path.resolve(options.directory, name);

  datastore_options.inMemoryOnly = options.inMemoryOnly || false;
  datastore_options.autoload = true;
  return Promise.promisifyAll(new DataStore(datastore_options));
}

UserDataStore.prototype.set_u2f_meta = function(userid, app_id, meta) {
  var newDocument = {};
  newDocument.userid = userid;
  newDocument.appid = app_id;
  newDocument.meta = meta;

  var filter = {};
  filter.userid = userid;
  filter.appid = app_id;

  return this._u2f_meta_collection.updateAsync(filter, newDocument, { upsert: true });
}

UserDataStore.prototype.get_u2f_meta = function(userid, app_id) {
  var filter = {};
  filter.userid = userid;
  filter.appid = app_id;

  return this._u2f_meta_collection.findOneAsync(filter);
}

UserDataStore.prototype.save_authentication_trace = function(userid, type, is_success) {
  var newDocument = {};
  newDocument.userid = userid;
  newDocument.date = new Date();
  newDocument.is_success = is_success;
  newDocument.type = type;

  return this._authentication_traces_collection.insertAsync(newDocument);
}

UserDataStore.prototype.get_last_authentication_traces = function(userid, type, is_success, count) {
  var query = {};
  query.userid = userid;
  query.type = type;
  query.is_success = is_success;

  var query = this._authentication_traces_collection.find(query)
    .sort({ date: -1 }).limit(count);
  var query_promisified = Promise.promisify(query.exec, { context: query });
  return query_promisified();
}

UserDataStore.prototype.issue_identity_check_token = function(userid, token, data, max_age) {
  var newDocument = {};
  newDocument.userid = userid;
  newDocument.token = token;
  newDocument.content = { userid: userid, data: data };
  newDocument.max_date = new Date(new Date().getTime() + max_age);

  return this._identity_check_tokens_collection.insertAsync(newDocument);
}

UserDataStore.prototype.consume_identity_check_token = function(token) {
  var query = {};
  query.token = token;
  var that = this;
  var doc_content;
  
  return this._identity_check_tokens_collection.findOneAsync(query)
  .then(function(doc) {
    if(!doc) {
      return Promise.reject('Registration token does not exist');
    }

    var max_date = doc.max_date;
    var current_date = new Date();
    if(current_date > max_date) {
      return Promise.reject('Registration token is not valid anymore');
    }

    doc_content = doc.content;
    return Promise.resolve();
  })
  .then(function() {
    return that._identity_check_tokens_collection.removeAsync(query);
  })
  .then(function() {
    return Promise.resolve(doc_content);
  })
}

UserDataStore.prototype.set_totp_secret = function(userid, secret) {
  var doc = {}
  doc.userid = userid;
  doc.secret = secret;

  var query = {};
  query.userid = userid;
  return this._totp_secret_collection.updateAsync(query, doc, { upsert: true });
}

UserDataStore.prototype.get_totp_secret = function(userid) {
  var query = {};
  query.userid = userid;
  return this._totp_secret_collection.findOneAsync(query);
}
