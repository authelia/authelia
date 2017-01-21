
module.exports = verify;

var objectPath = require('object-path');
var Promise = require('bluebird');

function verify_filter(req, res) {
  if(!objectPath.has(req, 'session.auth_session'))
    return Promise.reject('No auth_session variable');

  if(!objectPath.has(req, 'session.auth_session.first_factor'))
    return Promise.reject('No first factor variable');

  if(!objectPath.has(req, 'session.auth_session.second_factor'))
    return Promise.reject('No second factor variable');

  if(!req.session.auth_session.first_factor || 
     !req.session.auth_session.second_factor)
    return Promise.reject('First or second factor not validated');
 
  return Promise.resolve();
}

function verify(req, res) {
  console.log('Verify authentication');

  verify_filter(req, res)
  .then(function() {
    res.status(204);
    res.send();
  })
  .catch(function(err) {
    res.status(401);
    res.send();
  });
}

