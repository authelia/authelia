// Based on: https://github.com/gohugoio/hugoDocs/blob/master/_vendor/github.com/gohugoio/gohugoioTheme/assets/js/tabs.js

/**
 * Scripts which manages Code Toggle tabs.
 */

const customTabsToggle = (name, event, allTabs, allPanes) => {
  let targetKey;

  if (event.target) {
    event.preventDefault();
    targetKey = event.currentTarget.getAttribute(`data-toggle-${name}-tab`)
  } else {
    targetKey = event
  }

  if (window.localStorage) {
    window.localStorage.setItem(`${name}TabPref`, targetKey)
  }

  let i;

  for (i = 0; i < allTabs.length; i++) {
    allTabs[i].classList.remove('active');
    allPanes[i].classList.remove('active');
  }

  const selectedTabs = document.querySelectorAll(`[data-toggle-${name}-tab=${targetKey}]`);
  const selectedPanes = document.querySelectorAll(`[data-${name}-pane=${targetKey}]`);

  for (i = 0; i < selectedTabs.length; i++) {
    selectedTabs[i].classList.add('active');
    selectedPanes[i].classList.add('show', 'active');
  }
}

const customTabsToggleListener = (name, allTabs, allPanes) => {
  return (ev) => {
    customTabsToggle(name, ev, allTabs, allPanes);
  }
};

const customTabsStorageListener = (name) => {
  return (ev) => {
    if (ev.key !== `${name}TabPref`) {
      return;
    }

    if (ev.newValue && ev.newValue !== '') {
      customTabsToggle(name, ev.newValue);
    }
  }
};

const customTabsConfigure = (name) => {
  // Find all of the related tabs on the page.
  const allTabs = document.querySelectorAll(`[data-toggle-${name}-tab]`);
  const allPanes = document.querySelectorAll(`[data-${name}-pane]`);

  // If no tabs or panes exist skip everything.
  if (allTabs.length === 0 && allPanes.length === 0) {
    return;
  }

  // If the browser supports localStorage, setup localStorage elements.
  if (window.localStorage) {

    // If the preference value exists, make sure those tabs are selected.
    const value = window.localStorage.getItem(`${name}TabPref`);
    if (value) {
      customTabsToggle(name, value, allTabs, allPanes)
    }

    // Make sure we listen for storage events for changes to the specific storage key.
    window.addEventListener('storage', customTabsStorageListener(name));
  }

  // Create the listener used for click events.
  const clickListener = customTabsToggleListener(name, allTabs, allPanes);

  // Make sure each tab has the click event listener.
  for (let i = 0; i < allTabs.length; i++) {
    allTabs[i].addEventListener('click', clickListener)
  }
};

// Register the 'env' tab group listeners etc. on page load.
customTabsConfigure('env');
customTabsConfigure('session');
