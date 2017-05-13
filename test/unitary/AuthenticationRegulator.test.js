
import { AuthenticationRegulator } from "../../src/lib/AuthenticationRegulator";
import * as UserDataStore from "../../src/lib/user_data_store";
import * as DataStore from "nedb";
import * as MockDate from "mockdate";

var exceptions = require('../../src/lib/exceptions');

describe.only('test authentication regulator', function() {
  it('should mark 2 authentication and regulate (resolve)', function() {
    var options = {};
    options.inMemoryOnly = true;
    var data_store = new UserDataStore(DataStore, options);
    var regulator = new AuthenticationRegulator(data_store, 10);
    var user = 'user';

    return regulator.mark(user, false)
    .then(function() {
      return regulator.mark(user, true);
    })
    .then(function() {
      return regulator.regulate(user);
    });
  });

  it('should mark 3 authentications and regulate (reject)', function(done) {
    var options = {};
    options.inMemoryOnly = true;
    var data_store = new UserDataStore(DataStore, options);
    var regulator = new AuthenticationRegulator(data_store, 10);
    var user = 'user';

    regulator.mark(user, false)
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.regulate(user);
    })
    .catch(exceptions.AuthenticationRegulationError, function() {
      done();
    })
  });

  it('should mark 3 authentications and regulate (resolve)', function(done) {
    var options = {};
    options.inMemoryOnly = true;
    var data_store = new UserDataStore(DataStore, options);
    var regulator = new AuthenticationRegulator(data_store, 10);
    var user = 'user';

    MockDate.set('1/2/2000 00:00:00');
    regulator.mark(user, false)
    .then(function() {
      MockDate.set('1/2/2000 00:00:15');
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.mark(user, false);
    })
    .then(function() {
      return regulator.regulate(user);
    })
    .then(function() {
      done();
    })
  });
});
