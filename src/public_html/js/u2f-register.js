(function() {

params={};
location.search.replace(/[?&]+([^=&]+)=([^&]*)/gi,function(s,k,v){params[k]=v});

function finishRegister(url, responseData, fn) {
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

function startRegister(fn, timeout) {
  $.get('/authentication/2ndfactor/u2f/register_request', {}, null, 'json')
  .done(function(startRegisterResponse) {
    u2f.register(
      startRegisterResponse.appId,
      startRegisterResponse.registerRequests,
      startRegisterResponse.registeredKeys,
      function (response) {
        if (response.errorCode) {
          fn(response.errorCode);
        } else {
          finishRegister('/authentication/2ndfactor/u2f/register', response, fn);
        }
      },
      timeout 
    );
  });
}

function redirect() {
  var redirect_uri = '/authentication/login';
  if('redirect' in params) {
    redirect_uri = params['redirect'];
  }
  window.location.replace(redirect_uri);
}

function onRegisterSuccess() {
  redirect();
}

function onRegisterFailure(err) {
  $.notify('Problem authenticating with U2F.', 'error');
}

$(document).ready(function() {
  startRegister(function(err) {
    if(err) {
      onRegisterFailure(err);
      return;
    }
    onRegisterSuccess();
  }, 240);
});

})();
