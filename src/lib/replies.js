
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

function authentication_succeeded(res, username, token) {
  console.log('Reply: authentication succeeded');
  res.status(200);
  res.set({ 'X-Remote-User': username });
  res.send(token);
}

function already_authenticated(res, username) {
  console.log('Reply: already authenticated');
  res.status(204);
  res.set({ 'X-Remote-User': username });
  res.send();
}

