
module.exports = UserDataStore;

var Promise = require('bluebird');
var path = require('path');

function UserDataStore(DataStore, options) {
  var datastore_options = {};
  if(options.directory) 
    datastore_options.filename = path.resolve(options.directory, 'u2f_meta');

  datastore_options.inMemoryOnly = options.inMemoryOnly || false;
  datastore_options.autoload = true;
  console.log(datastore_options);

  this._u2f_meta_collection = Promise.promisifyAll(new DataStore(datastore_options));
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

