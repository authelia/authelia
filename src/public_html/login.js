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
      onFirstFactorFailure();
      return;
    }
    onFirstFactorSuccess();
  });
}

function onTotpSignButtonClicked() {
  var token = $('#totp-token').val();
  validateSecondFactorTotp(token, function(err) {
    if(err) {
      onSecondFactorTotpFailure();
      return;
    }
    onSecondFactorTotpSuccess();
  });
}

function onU2fSignButtonClicked() {
  startSecondFactorU2fSigning(function(err) {
    if(err) {
      onSecondFactorU2fSigningFailure();
      return;
    }
    onSecondFactorU2fSigningSuccess();
  }, 120);
}

function onU2fRegisterButtonClicked() {
  startSecondFactorU2fRegister(function(err) {
    if(err) {
      onSecondFactorU2fRegisterFailure();
      return;
    }
    onSecondFactorU2fRegisterSuccess();
  }, 120);
}

function finishSecondFactorU2f(url, responseData, fn) {
  console.log(responseData);
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
    $.notify('Error when finish U2F transaction' + status);
  });
}

function startSecondFactorU2fSigning(fn, timeout) {
  $.get('/auth/2ndfactor/u2f/sign_request', {}, null, 'json')
  .done(function(signResponse) {
    var registeredKeys = signResponse.registeredKeys;
    $.notify('Please touch the token', 'information');
    console.log(signResponse);

    // Store sessionIds
    // var sessionIds = {};
    // for (var i = 0; i < registeredKeys.length; i++) {
    //   sessionIds[registeredKeys[i].keyHandle] = registeredKeys[i].sessionId;
    //   delete registeredKeys[i]['sessionId'];
    // }

    u2f.sign(
      signResponse.appId,
      signResponse.challenge,
      signResponse.registeredKeys,
      function (response) {
        if (response.errorCode) {
          fn(response);
        } else {
          // response['sessionId'] = sessionIds[response.keyHandle];
          finishSecondFactorU2f('/auth/2ndfactor/u2f/sign', response, fn);
        }
      },
      timeout
    );
  })
  .fail(function(xhr, status) {
     fn(status);
  });
}

function startSecondFactorU2fRegister(fn, timeout) {
  $.get('/auth/2ndfactor/u2f/register_request', {}, null, 'json')
  .done(function(startRegisterResponse) {
    console.log(startRegisterResponse);
    $.notify('Please touch the token', 'information');
    u2f.register(
      startRegisterResponse.appId,
      startRegisterResponse.registerRequests,
      startRegisterResponse.registeredKeys,
      function (response) {
        if (response.errorCode) {
          fn(response.errorCode);
        } else {
          // response['sessionId'] = startRegisterResponse.clientData;
          finishSecondFactorU2f('/auth/2ndfactor/u2f/register', response, fn);
        }
      },
      timeout 
    );
  });
}

function validateSecondFactorTotp(token, fn) {
  $.post('/auth/2ndfactor/totp', {
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
  $.post('/auth/1stfactor', {
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

function onFirstFactorFailure() {
  $('#password').val('');
  $('#token').val('');
  $.notify('Wrong credentials', 'error');
}

function onAuthenticationSuccess() {
  $.notify('Authentication succeeded. You are redirected.', 'success');
  redirect();
}

function onSecondFactorTotpSuccess() {
  onAuthenticationSuccess();
}

function onSecondFactorTotpFailure() {
  $.notify('Wrong TOTP token', 'error');
}

function onSecondFactorU2fSigningSuccess() {
  onAuthenticationSuccess();
}

function onSecondFactorU2fSigningFailure(err) {
  console.error(err);
  $.notify('Problem authenticating with U2F.', 'error');
}

function onSecondFactorU2fRegisterSuccess() {
  $.notify('Registration succeeded. You can now sign in.', 'success');
}

function onSecondFactorU2fRegisterFailure(err) {
  console.error(err);
  $.notify('Problem authenticating with U2F.', 'error');
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
  $('#first-factor #information').hide();
}

function cleanupFirstFactorLoginButton() {
  $('#first-factor #login-button').off('click');
}

function setupTotpSignButton() {
  $('#second-factor #totp-sign-button').on('click', onTotpSignButtonClicked);
  setupEnterKeypressListener('#totp', onTotpSignButtonClicked);
}

function setupU2fSignButton() {
  $('#second-factor #u2f-sign-button').on('click', onU2fSignButtonClicked);
  setupEnterKeypressListener('#u2f', onU2fSignButtonClicked);
}

function setupU2fRegisterButton() {
  $('#second-factor #u2f-register-button').on('click', onU2fRegisterButtonClicked);
  setupEnterKeypressListener('#u2f', onU2fRegisterButtonClicked);
}

function enterFirstFactor() {
  // console.log('entering first factor');
  showFirstFactorLayout();
  hideSecondFactorLayout();
  setupFirstFactorLoginButton();
}

function enterSecondFactor() {
  // console.log('entering second factor');
  hideFirstFactorLayout();
  showSecondFactorLayout();
  cleanupFirstFactorLoginButton();
  setupTotpSignButton();
  setupU2fSignButton();
  setupU2fRegisterButton();
}

$(document).ready(function() {
  enterFirstFactor();
});

})();
