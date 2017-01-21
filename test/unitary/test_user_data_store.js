
var UserDataStore = require('../../src/lib/user_data_store');
var DataStore = require('nedb');
var assert = require('assert');
var Promise = require('bluebird');

describe('test user data store', function() {
  it('should save a u2f meta', function() {
    var options = {};
    options.inMemoryOnly = true;

    var data_store = new UserDataStore(DataStore, options);

    var userid = 'user';
    var app_id = 'https://localhost';
    var meta = {};
    meta.publicKey = 'pbk';

    return data_store.set_u2f_meta(userid, app_id, meta)
    .then(function(numUpdated) {
      assert.equal(1, numUpdated);
      return Promise.resolve();
    });
  });

  it('should retrieve no u2f meta', function() {
    var options = {};
    options.inMemoryOnly = true;

    var data_store = new UserDataStore(DataStore, options);

    var userid = 'user';
    var app_id = 'https://localhost';
    var meta = {};
    meta.publicKey = 'pbk';

    return data_store.get_u2f_meta(userid, app_id)
    .then(function(doc) {
      assert.equal(undefined, doc);
      return Promise.resolve();
    });
  });

  it('should insert and retrieve a u2f meta', function() {
    var options = {};
    options.inMemoryOnly = true;

    var data_store = new UserDataStore(DataStore, options);

    var userid = 'user';
    var app_id = 'https://localhost';
    var meta = {};
    meta.publicKey = 'pbk';

    return data_store.set_u2f_meta(userid, app_id, meta)
    .then(function(numUpdated, data) {
      assert.equal(1, numUpdated);
      return data_store.get_u2f_meta(userid, app_id)
    })
    .then(function(doc) {
      assert.deepEqual(meta, doc.meta);
      assert.deepEqual(userid, doc.userid);
      assert.deepEqual(app_id, doc.appid);
      assert('_id' in doc);
      return Promise.resolve();
    });
  });
});
