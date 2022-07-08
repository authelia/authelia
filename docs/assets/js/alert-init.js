Object.keys(localStorage).forEach(function(key) {
  if (/^global-alert-/.test(key)) {
    document.documentElement.setAttribute('data-global-alert', 'closed');
  }
});