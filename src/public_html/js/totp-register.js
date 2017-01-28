(function() {

function generateSecret(fn) {
  $.ajax({
    type: 'POST',
    url: '/authentication/new-totp-secret',
    contentType: 'application/json',
    dataType: 'json',
  })
  .done(function(data) {
    fn(undefined, data);
  })
  .fail(function(xhr, status) {
    $.notify('Error when generating TOTP secret');
  });
}

function onSecretGenerated(err, secret) {
  // console.log('secret generated successfully', secret);
  var img = $('<img src="' + secret.qrcode + '" alt="secret-qrcode"/>');
  $('#qrcode').append(img);
  $("#secret").text(secret.base32);
}

$(document).ready(function() {
  generateSecret(onSecretGenerated);
});
})();
