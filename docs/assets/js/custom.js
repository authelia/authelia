// Based on: https://github.com/gohugoio/hugoDocs/blob/master/_vendor/github.com/gohugoio/gohugoioTheme/assets/js/tabs.js

/**
 * Scripts which manages Code Toggle tabs.
 */
// store tabs variable
var allEnvTabs = document.querySelectorAll('[data-toggle-env-tab]');
var allEnvPanes = document.querySelectorAll('[data-env-pane]');
const localStorageKeyEnvTabs = 'envPref';

function toggleEnvTabs(event) {

  if(event.target){
    event.preventDefault();
    var clickedTab = event.currentTarget;
    var targetKey = clickedTab.getAttribute('data-toggle-env-tab')
  } else {
    var targetKey = event
  }
  // We store the config language selected in users' localStorage
  if(window.localStorage){
    window.localStorage.setItem(localStorageKeyEnvTabs, targetKey)
  }
  var selectedTabs = document.querySelectorAll('[data-toggle-env-tab=' + targetKey + ']');
  var selectedPanes = document.querySelectorAll('[data-env-pane=' + targetKey + ']');

  for (var i = 0; i < allEnvTabs.length; i++) {
    allEnvTabs[i].classList.remove('active');
    allEnvPanes[i].classList.remove('active');
  }

  for (var i = 0; i < selectedTabs.length; i++) {
    selectedTabs[i].classList.add('active');
    selectedPanes[i].classList.add('show', 'active');
  }

}

const envTabsStorageListener = (ev) => {
  if (ev.key !== localStorageKeyEnvTabs) {
    return;
  }

  if (ev.newValue && ev.newValue !== '') {
    toggleEnvTabs(ev.newValue);
  }
};

for (var i = 0; i < allEnvTabs.length; i++) {
  allEnvTabs[i].addEventListener('click', toggleEnvTabs)
}

window.addEventListener('storage', envTabsStorageListener);

// Upon page load, if user has a preferred language in its localStorage, tabs are set to it.
if(window.localStorage.getItem(localStorageKeyEnvTabs)) {
  toggleEnvTabs(window.localStorage.getItem(localStorageKeyEnvTabs))
}
