
var assert = require('assert');
var Promise = require('bluebird');
var sinon = require('sinon');
var MockDate = require('mockdate');
var UserDataStore = require('../../../src/lib/user_data_store');
var DataStore = require('nedb');

describe('test user data store', function() {
  describe('test totp secrets store', test_totp_secrets);
});

function test_totp_secrets() {
  it('should save and reload a totp secret', function() {
    var options = {};
    options.inMemoryOnly = true;

    var data_store = new UserDataStore(DataStore, options);
    var userid = 'user';
    var secret = {};
    secret.ascii = 'abc';
    secret.base32 = 'ABCDKZLEFZGREJK';

    return data_store.set_totp_secret(userid, secret)
    .then(function() {
      return data_store.get_totp_secret(userid);
    })
    .then(function(doc) {
      assert('_id' in doc);
      assert.equal(doc.userid, 'user');
      assert.equal(doc.secret.ascii, 'abc');
      assert.equal(doc.secret.base32, 'ABCDKZLEFZGREJK');
      return Promise.resolve();
    });
  });

  it('should only remember last secret', function() {
    var options = {};
    options.inMemoryOnly = true;

    var data_store = new UserDataStore(DataStore, options);
    var userid = 'user';
    var secret1 = {};
    secret1.ascii = 'abc';
    secret1.base32 = 'ABCDKZLEFZGREJK';
    var secret2 = {};
    secret2.ascii = 'def';
    secret2.base32 = 'XYZABC';

    return data_store.set_totp_secret(userid, secret1)
    .then(function() {
      return data_store.set_totp_secret(userid, secret2)
    })
    .then(function() {
      return data_store.get_totp_secret(userid);
    })
    .then(function(doc) {
      assert('_id' in doc);
      assert.equal(doc.userid, 'user');
      assert.equal(doc.secret.ascii, 'def');
      assert.equal(doc.secret.base32, 'XYZABC');
      return Promise.resolve();
    });
  });
}
