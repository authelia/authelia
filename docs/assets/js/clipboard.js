import Clipboard from 'clipboard';

var pre = document.getElementsByTagName('pre');

for (var i = 0; i < pre.length; ++ i)
{
  var element = pre[i];
  var mermaid = element.getElementsByClassName('language-mermaid')[0];

  if (mermaid == null) {
    element.insertAdjacentHTML('afterbegin', '<button class="btn btn-copy"></button>');
  }
}

var clipboard = new Clipboard('.btn-copy', {

  target: function(trigger) {
    return trigger.nextElementSibling;
  },

});

clipboard.on('success', function(e) {

    /*
    console.info('Action:', e.action);
    console.info('Text:', e.text);
    console.info('Trigger:', e.trigger);
    */

    e.clearSelection();
});

clipboard.on('error', function(e) {
    console.error('Action:', e.action);
    console.error('Trigger:', e.trigger);
});
