(function() {

function setupEnterKeypressListener(filter, fn) {
  $(filter).on('keydown', 'input', function (e) {
    var key = e.which;
    switch (key) {
    case 13: // enter key code
      fn();
      break;
    default:
      break;
    }
  });
}

function onResetPasswordButtonClicked() {
  var username = $('#username').val();
 
  if(!username) {
    $.notify('You must provide your username to reset your password.', 'warn');
    return;
  }
  
  $.post('/authentication/reset-password', {
    userid: username,
  })
  .done(function() {
    $.notify('An email has been sent. Click on the link to change your password', 'success');
    setTimeout(function() {
      window.location.replace('/authentication/login');
    }, 1000);
  })
  .fail(function() {
    $.notify('Are you sure this is your username?', 'warn');
  });
}

function setupResetPasswordButton() {
  $('#reset-password-button').on('click', onResetPasswordButtonClicked);
}

$(document).ready(function() {
  setupResetPasswordButton();
  setupEnterKeypressListener('#reset-password-form', onResetPasswordButtonClicked);
});

})();
