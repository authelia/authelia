(function() {

params={};
location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(s,k,v){params[k]=v});
console.log(params);

$(document).ready(function() {
  $('#login-button').on('click', onLoginButtonClicked);
  setupEnterKeypressListener();
  $('#information').hide();
});

function setupEnterKeypressListener() {
  $('#login-form').on('keydown', 'input', function (e) {
    var key = e.which;
    switch (key) {
    case 13: // enter key code
      onLoginButtonClicked();
      break;
    default:
      break;
    }
  });
}

function onLoginButtonClicked() {
  var username = $('#username').val();
  var password = $('#password').val();
  var token = $('#token').val();
  
  authenticate(username, password, token, function(err, access_token) {
    if(err) {
      onAuthenticationFailure();
      return;
    }
    onAuthenticationSuccess(access_token);
  });
}


function authenticate(username, password, token, fn) {
  $.post('/_auth', {
    username: username,
    password: password,
    token: token
  })
  .done(function(access_token) {
    fn(undefined, access_token);
  })
  .fail(function(err) {
    fn(err);
  });
}

function displayInformationMessage(msg, type, time, fn) {
  if(type == 'success') {
    $('#information').addClass("success");
  }
  else if(type == 'failure') {
    $('#information').addClass("failure");
  }
  
  $('#information').text(msg);
  $('#information').show("fast");

  setTimeout(function() {
    $('#information').hide("fast");
    $('#information').removeClass("success");
    $('#information').removeClass("failure");

    if(fn) fn();
  },time);
}

function redirect() {
  var redirect_uri = '/';
  if('redirect' in params) {
    redirect_uri = params['redirect'];
  }

  window.location.replace(redirect_uri);
}

function onAuthenticationSuccess(access_token) {
  Cookies.set('access_token', access_token, { path: '/' });

  $('#username').val('');
  $('#password').val('');
  $('#token').val('');
 
  redirect();
  // displayInformationMessage('Authentication success, You will be redirected' +
  //                           'in few seconds.', 'success', 3000, function() {
  // });
}

function onAuthenticationFailure() {
  $('#password').val('');
  $('#token').val('');

  displayInformationMessage('Authentication failed, please try again.', 'failure', 3000);
}

})();
