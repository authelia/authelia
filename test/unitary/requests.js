
var Promise = require('bluebird');
var request = Promise.promisifyAll(require('request'));
var assert = require('assert');

module.exports = function(port) {
  var PORT = port;
  var BASE_URL = 'http://localhost:' + PORT;

  function execute_reset_password(jar, transporter, user, new_password) {
    return request.postAsync({
      url: BASE_URL + '/authentication/reset-password',
      jar: jar,
      form: { userid: user }
    })
    .then(function(res) {
      assert.equal(res.statusCode, 204);
      var html_content = transporter.sendMail.getCall(0).args[0].html;
      var regexp = /identity_token=([a-zA-Z0-9]+)/;
      var token = regexp.exec(html_content)[1];
      // console.log(html_content, token);
      return request.getAsync({
        url: BASE_URL + '/authentication/reset-password?identity_token=' + token,
        jar: jar
      })
    })
    .then(function(res) {
      assert.equal(res.statusCode, 200); 
      return request.postAsync({
        url: BASE_URL + '/authentication/new-password',
        jar: jar,
        form: {
          password: new_password
        }
      });
    });
  }

  function execute_totp(jar, token) {
    return request.postAsync({
      url: BASE_URL + '/authentication/2ndfactor/totp',
      jar: jar,
      form: {
        token: token
      }
    });
  }
  
  function execute_u2f_authentication(jar) {
    return request.getAsync({
      url: BASE_URL + '/authentication/2ndfactor/u2f/sign_request',
      jar: jar
    })
    .then(function(res) {
      assert.equal(res.statusCode, 200); 
      return request.postAsync({
        url: BASE_URL + '/authentication/2ndfactor/u2f/sign',
        jar: jar,
        form: {
        }
      });
    });
  }
  
  function execute_verification(jar) {
    return request.getAsync({ url: BASE_URL + '/authentication/verify', jar: jar })
  }
  
  function execute_login(jar) {
    return request.getAsync({ url: BASE_URL + '/authentication/login', jar: jar })
  }
  
  function execute_u2f_registration(jar, transporter) {
    return request.postAsync({
      url: BASE_URL + '/authentication/u2f-register',
      jar: jar
    })
    .then(function(res) {
      assert.equal(res.statusCode, 204);
      var html_content = transporter.sendMail.getCall(0).args[0].html;
      var regexp = /identity_token=([a-zA-Z0-9]+)/;
      var token = regexp.exec(html_content)[1];
      // console.log(html_content, token);
      return request.getAsync({
        url: BASE_URL + '/authentication/u2f-register?identity_token=' + token,
        jar: jar
      })
    })
    .then(function(res) {
      assert.equal(res.statusCode, 200); 
      return request.getAsync({
        url: BASE_URL + '/authentication/2ndfactor/u2f/register_request',
        jar: jar,
      });
    })
    .then(function(res) {
      assert.equal(res.statusCode, 200); 
      return request.postAsync({
        url: BASE_URL + '/authentication/2ndfactor/u2f/register',
        jar: jar,
        form: {
          s: 'test'
        }
      });
    });
  }
  
  function execute_first_factor(jar) {
    return request.postAsync({ 
      url: BASE_URL + '/authentication/1stfactor',
      jar: jar,
      form: {
        username: 'test_ok',
        password: 'password'
      }
    });
  }

  function execute_failing_first_factor(jar) {
    return request.postAsync({ 
      url: BASE_URL + '/authentication/1stfactor',
      jar: jar,
      form: {
        username: 'test_nok',
        password: 'password'
      }
    });
  }
  
  return {
    login: execute_login,
    verify: execute_verification,  
    reset_password: execute_reset_password,
    u2f_authentication: execute_u2f_authentication,
    u2f_registration: execute_u2f_registration,
    first_factor: execute_first_factor,
    failing_first_factor: execute_failing_first_factor,
    totp: execute_totp,
  }

}

