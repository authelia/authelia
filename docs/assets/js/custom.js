// Based on: https://github.com/gohugoio/hugoDocs/blob/master/_vendor/github.com/gohugoio/gohugoioTheme/assets/js/tabs.js

/**
 * Scripts which manages Code Toggle tabs.
 */

const customTabsStorageName = (name) => {
  return `tab-preference-${name}`;
};

const customTabsToggle = (name, event, allTabs, allPanes) => {
  let targetKey;

  if (event.target) {
    event.preventDefault();
    targetKey = event.currentTarget.getAttribute(`data-toggle-${name}-tab`)
  } else {
    targetKey = event
  }

  if (window.localStorage) {
    window.localStorage.setItem(customTabsStorageName(name), targetKey)
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
    if (ev.key !== customTabsStorageName(name)) {
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
    const value = window.localStorage.getItem(customTabsStorageName(name));
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

const siteVariableName = (name) => {
  return `site-variable-${name}`;
};

const siteVariableReplace = (name, value) => {
  const standard= document.getElementsByClassName(siteVariableName(name));

  [].slice.call(standard).forEach((item) => {
    item.innerHTML = value;
  });

  if (name === "domain") {
    siteVariableReplaceDomain(value);
  }
};

const siteVariableReplaceDomain = (value) => {
  const itemsRegex= document.getElementsByClassName(siteVariableName("domain")+"-regex");

  [].slice.call(itemsRegex).forEach((item) => {
    item.innerHTML = value.replace(".", "\\.");
  });

  const itemsDN= document.getElementsByClassName(siteVariableName("domain")+"-dn");

  [].slice.call(itemsDN).forEach((item) => {
    item.innerHTML = `DC=${value.replace(".", ",DC=")}`;
  });
};

const siteVariableStorageListener = (name) => {
  return (ev) => {
    if (ev.key !== siteVariableName(name)) {
      return;
    }

    if (ev.newValue && ev.newValue !== '') {
      siteVariableReplace(name, ev.newValue);
    }
  }
};

const siteVariableConfigure = (name, fallback) => {
  var finalValue = fallback;

  // If the browser supports localStorage, setup localStorage elements.
  if (window.localStorage) {

    // If the preference value exists, make sure those tabs are selected.
    const value = window.localStorage.getItem(siteVariableName(name));
    if (value && value !== "") {
      finalValue = value;
      siteVariableReplace(name, value)
    } else {
      siteVariableReplace(name, fallback)
    }

    // Make sure we listen for storage events for changes to the specific storage key.
    window.addEventListener('storage', siteVariableStorageListener(name));
  } else {
    siteVariableReplace(name, fallback);
  }

  return finalValue;
};

const siteVariableSet = (name, value, prev) => {
  if (value === prev) {
    return prev;
  }

  siteVariableReplace(name, value);

  if (window.localStorage) {
    window.localStorage.setItem(siteVariableName(name), value);
  }

  return value;
};

const siteVariablesConfigure = () => {
  var domain = siteVariableConfigure("domain", "example.com");
  var subdomainAuthelia = siteVariableConfigure("subdomain-authelia", "auth");

  const save = document.getElementById("site-variables-save");
  if (!save) return;

  const onChangeAutheliaDomain = () => {
    const valueDomain = document.getElementById(siteVariableName("domain")).value.trim();
    const valueSubdomain = document.getElementById(siteVariableName("subdomain-authelia")).value.trim();

    document.getElementById("site-const-authelia-url").value = `https://${valueSubdomain}.${valueDomain}/`;
  };

  document.getElementById("site-variables-toggle").addEventListener("click", () => {
    document.getElementById(siteVariableName("domain")).value = domain;
    document.getElementById(siteVariableName("subdomain-authelia")).value = subdomainAuthelia;
    onChangeAutheliaDomain();
  })

  save.addEventListener("click", () => {
    domain = siteVariableSet("domain", document.getElementById(siteVariableName("domain")).value.trim(), domain);
    subdomainAuthelia = siteVariableSet("subdomain-authelia", document.getElementById(siteVariableName("subdomain-authelia")).value.trim(), subdomainAuthelia);
  })

  document.getElementById("site-variable-domain").addEventListener("change", onChangeAutheliaDomain);
  document.getElementById("site-variable-domain").addEventListener("keyup", onChangeAutheliaDomain);
  document.getElementById("site-variable-subdomain-authelia").addEventListener("change", onChangeAutheliaDomain);
  document.getElementById("site-variable-subdomain-authelia").addEventListener("keyup", onChangeAutheliaDomain);
};

// Register the 'env' tab group listeners etc. on page load.
customTabsConfigure('env');
customTabsConfigure('session');

siteVariablesConfigure();
