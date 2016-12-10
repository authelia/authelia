
module.exports = {
  'authentication_failed': authentication_failed,
  'authentication_succeeded': authentication_succeeded,
  'already_authenticated': already_authenticated
}

function authentication_failed(res) {
  console.log('Reply: authentication failed');
  res.status(401)
  res.send('Authentication failed');
}

function send_success(res, username, msg) {
  res.status(200);
  res.set({ 'X-Remote-User': username });
  res.send(msg);
}

function authentication_succeeded(res, username) {
  console.log('Reply: authentication succeeded');
  send_success(res, username, 'Authentication succeeded');
}

function already_authenticated(res, username) {
  console.log('Reply: already authenticated');
  send_success(res, username, 'Authentication succeeded');
}

