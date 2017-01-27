
var objectPath = require('object-path');
var Promise = require('bluebird');

var CHALLENGE = 'u2f-register';

var icheck_interface = {
  challenge: CHALLENGE,
  render_template: 'u2f-register',
  pre_check_callback: pre_check,
}

module.exports = {
  icheck_interface: icheck_interface,
}


function pre_check(req) {
  var first_factor_passed = objectPath.get(req, 'session.auth_session.first_factor');
  if(!first_factor_passed) {
    return Promise.reject('Authentication required before issuing a u2f registration request');
  }

  var userid = objectPath.get(req, 'session.auth_session.userid');
  var email = objectPath.get(req, 'session.auth_session.email');

  if(!(userid && email)) {
    return Promise.reject('User ID or email is missing');
  }

  var identity = {};
  identity.email = email;
  identity.userid = userid;
  return Promise.resolve(identity);
}

