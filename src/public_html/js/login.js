(function() {

params={};
location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(s,k,v){params[k]=v});


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

function onLoginButtonClicked() {
  var username = $('#username').val();
  var password = $('#password').val();

  validateFirstFactor(username, password, function(err) {
    if(err) {
      onFirstFactorFailure(err.responseText);
      return;
    }
    onFirstFactorSuccess();
  });
}

function onResetPasswordButtonClicked() {
  var r = '/authentication/reset-password-form';
  window.location.replace(r);
}

function onTotpSignButtonClicked() {
  var token = $('#totp-token').val();
  validateSecondFactorTotp(token, function(err) {
    if(err) {
      onSecondFactorTotpFailure(err.responseText);
      return;
    }
    onSecondFactorTotpSuccess();
  });
}

function onTotpRegisterButtonClicked() {
  $.ajax({
    type: 'POST',
    url: '/authentication/totp-register'
  })
  .done(function(data) {
    $.notify('An email has been sent to your email address', 'info');
  })
  .fail(function(xhr, status) {
    $.notify('Unable to send you an email', 'error');
  });
}

function onU2fSignButtonClicked() {
  startU2fAuthentication(function(err) {
    if(err) {
      onU2fAuthenticationFailure();
      return;
    }
    onU2fAuthenticationSuccess();
  }, 120);
}

function onU2fRegistrationButtonClicked() {
  askForU2fRegistration(function(err) {
    if(err) {
      $.notify('Unable to send you an email', 'error');
      return;
    }
    $.notify('An email has been sent to your email address', 'info');
  });
}

function askForU2fRegistration(fn) {
  $.ajax({
    type: 'POST',
    url: '/authentication/u2f-register'
  })
  .done(function(data) {
    fn(undefined, data);
  })
  .fail(function(xhr, status) {
    fn(status);
  });
}

function finishU2fAuthentication(url, responseData, fn) {
  $.ajax({
    type: 'POST',
    url: url,
    data: JSON.stringify(responseData),
    contentType: 'application/json',
    dataType: 'json',
  })
  .done(function(data) {
    fn(undefined, data);
  })
  .fail(function(xhr, status) {
    $.notify('Error when finish U2F transaction', 'error');
  });
}

function startU2fAuthentication(fn, timeout) {
  $.get('/authentication/2ndfactor/u2f/sign_request', {}, null, 'json')
  .done(function(signResponse) {
    var registeredKeys = signResponse.registeredKeys;
    $.notify('Please touch the token', 'info');

    u2f.sign(
      signResponse.appId,
      signResponse.challenge,
      signResponse.registeredKeys,
      function (response) {
        if (response.errorCode) {
          fn(response);
        } else {
          finishU2fAuthentication('/authentication/2ndfactor/u2f/sign', response, fn);
        }
      },
      timeout
    );
  })
  .fail(function(xhr, status) {
     fn(status);
  });
}

function validateSecondFactorTotp(token, fn) {
  $.post('/authentication/2ndfactor/totp', {
    token: token,
  })
  .done(function() {
    fn(undefined);
  })
  .fail(function(err) {
    fn(err);
  });
}

function validateFirstFactor(username, password, fn) {
  $.post('/authentication/1stfactor', {
    username: username,
    password: password,
  })
  .done(function() {
    fn(undefined);
  })
  .fail(function(err) {
    fn(err);
  });
}

function redirect() {
  var redirect_uri = '/';
  if('redirect' in params) {
    redirect_uri = params['redirect'];
  }
  window.location.replace(redirect_uri);
}

function onFirstFactorSuccess() {
  $('#username').val('');
  $('#password').val('');
  enterSecondFactor();
}

function onFirstFactorFailure(err) {
  $('#password').val('');
  $('#token').val('');
  $.notify('Error during authentication: ' + err, 'error');
}

function onAuthenticationSuccess() {
  $.notify('Authentication succeeded. You are redirected.', 'success');
  redirect();
}

function onSecondFactorTotpSuccess() {
  onAuthenticationSuccess();
}

function onSecondFactorTotpFailure(err) {
  $.notify('Error while validating TOTP token. Cause: ' + err, 'error');
}

function onU2fAuthenticationSuccess() {
  onAuthenticationSuccess();
}

function onU2fAuthenticationFailure(err) {
  $.notify('Problem with U2F authentication. Did you register before authenticating?', 'warn');
}

function showFirstFactorLayout() {
  $('#first-factor').show();
}

function hideFirstFactorLayout() {
  $('#first-factor').hide();
}

function showSecondFactorLayout() {
  $('#second-factor').show();
}

function hideSecondFactorLayout() {
  $('#second-factor').hide();
}

function setupFirstFactorLoginButton() {
  $('#first-factor #login-button').on('click', onLoginButtonClicked);
  setupEnterKeypressListener('#login-form', onLoginButtonClicked);
}

function cleanupFirstFactorLoginButton() {
  $('#first-factor #login-button').off('click');
}

function setupTotpSignButton() {
  $('#second-factor #totp-sign-button').on('click', onTotpSignButtonClicked);
  setupEnterKeypressListener('#totp', onTotpSignButtonClicked);
}

function setupTotpRegisterButton() {
  $('#second-factor #totp-register-button').on('click', onTotpRegisterButtonClicked);
}

function setupU2fSignButton() {
  $('#second-factor #u2f-sign-button').on('click', onU2fSignButtonClicked);
  setupEnterKeypressListener('#u2f', onU2fSignButtonClicked);
}

function setupU2fRegistrationButton() {
  $('#second-factor #u2f-register-button').on('click', onU2fRegistrationButtonClicked);
}

function setupResetPasswordButton() {
  $('#first-factor #reset-password-button').on('click', onResetPasswordButtonClicked);
}

function enterFirstFactor() {
  showFirstFactorLayout();
  hideSecondFactorLayout();
  setupFirstFactorLoginButton();
  setupResetPasswordButton();
}

function enterSecondFactor() {
  hideFirstFactorLayout();
  showSecondFactorLayout();
  cleanupFirstFactorLoginButton();
  setupTotpSignButton();
  setupTotpRegisterButton();
  setupU2fSignButton();
  setupU2fRegistrationButton();
}

$(document).ready(function() {
  enterFirstFactor();
});

})();
