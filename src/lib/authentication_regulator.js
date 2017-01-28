
module.exports = AuthenticationRegulator;

var exceptions = require('./exceptions');
var Promise = require('bluebird');

function AuthenticationRegulator(user_data_store, lock_time_in_seconds) {
  this._user_data_store = user_data_store;
  this._lock_time_in_seconds = lock_time_in_seconds;
}

// Mark authentication
AuthenticationRegulator.prototype.mark = function(userid, is_success) {
  return this._user_data_store.save_authentication_trace(userid, '1stfactor', is_success);
}

AuthenticationRegulator.prototype.regulate = function(userid) {
  var that = this;
  return this._user_data_store.get_last_authentication_traces(userid, '1stfactor', false, 3)
  .then(function(docs) {
    if(docs.length < 3) {
      return Promise.resolve();
    }

    var oldest_doc = docs[2];
    var no_lock_min_date = new Date(new Date().getTime() - 
                                    that._lock_time_in_seconds * 1000);
    
    if(oldest_doc.date > no_lock_min_date) {
      throw new exceptions.AuthenticationRegulationError();
    }

    return Promise.resolve();
  });
}
