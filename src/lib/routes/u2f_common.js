
module.exports = {
  extract_app_id: extract_app_id,
  extract_original_url: extract_original_url,
  extract_referrer: extract_referrer,
  reply_with_internal_error: reply_with_internal_error,
  reply_with_missing_registration: reply_with_missing_registration,
  reply_with_unauthorized: reply_with_unauthorized
}

var util = require('util');

function extract_app_id(req) {
  return util.format('https://%s', req.headers.host);
}

function extract_original_url(req) {
  return util.format('https://%s%s', req.headers.host, req.headers['x-original-uri']);
}

function extract_referrer(req) {
  return req.headers.referrer;
}

function reply_with_internal_error(res, msg) {
  res.status(500);
  res.send(msg)
}

function reply_with_missing_registration(res) {
  res.status(401);
  res.send('Please register before authenticate');
}

function reply_with_unauthorized(res) {
  res.status(401);
  res.send();
}
