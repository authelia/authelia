
module.exports = totp_fn;

var objectPath = require('object-path');
var exceptions = require('../../../src/lib/Exceptions');

var UNAUTHORIZED_MESSAGE = 'Unauthorized access';

function totp_fn(req, res) {
  var logger = req.app.get('logger');
  var userid = objectPath.get(req, 'session.auth_session.userid');
  logger.info('POST 2ndfactor totp: Initiate TOTP validation for user %s', userid);

  if(!userid) {
    logger.error('POST 2ndfactor totp: No user id in the session');
    res.status(403);
    res.send();
    return;
  }

  var token = req.body.token;
  var totpValidator = req.app.get('totp validator');
  var data_store = req.app.get('user data store');  

  logger.debug('POST 2ndfactor totp: Fetching secret for user %s', userid);
  data_store.get_totp_secret(userid)
  .then(function(doc) {
    logger.debug('POST 2ndfactor totp: TOTP secret is %s', JSON.stringify(doc));
    return totpValidator.validate(token, doc.secret.base32);
  })
  .then(function() {
    logger.debug('POST 2ndfactor totp: TOTP validation succeeded');
    objectPath.set(req, 'session.auth_session.second_factor', true);
    res.status(204);
    res.send();
  }, function(err) {
    throw new exceptions.InvalidTOTPError();
  })
  .catch(exceptions.InvalidTOTPError, function(err) {
    logger.error('POST 2ndfactor totp: Invalid TOTP token %s', err);
    res.status(401);
    res.send('Invalid TOTP token');
  })
  .catch(function(err) {
    logger.error('POST 2ndfactor totp: Internal error %s', err);
    res.status(500);
    res.send('Internal error');
  });
}
