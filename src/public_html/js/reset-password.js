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
  var password1 = $('#password1').val();
  var password2 = $('#password2').val();
 
  if(!password1 || !password2) {
    $.notify('You must enter your new password twice.', 'warn');
    return;
  }

  if(password1 != password2) {
    $.notify('The passwords are different', 'warn');
    return;
  }
  
  $.post('/new-password', {
    password: password1,
  })
  .done(function() {
    $.notify('Your password has been changed. Please login again', 'success');
    window.location.replace('/login');
  })
  .fail(function() {
    $.notify('An error occurred during password change.', 'warn');
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
