// Based on: https://github.com/gohugoio/hugoDocs/blob/master/_vendor/github.com/gohugoio/gohugoioTheme/assets/js/tabs.js
// Put your custom JS code here
import { Popover } from 'bootstrap';

const popoverTriggerList = document.querySelectorAll('[data-bs-toggle="popover"]')
const popoverList = [...popoverTriggerList].map(popoverTriggerEl => new Popover(popoverTriggerEl))

const variables = {
  "host": {
    "type": "string",
    "value": "",
    "fallback": "authelia",
  },
  "port": {
    "type": "number",
    "value": 0,
    "fallback": 9091,
  },
  "tls": {
    "type": "boolean",
    "value": false,
    "false": "http",
    "true": "https",
    "fallback": false,
  },
  "domain": {
    "type": "string",
    "value": "",
    "fallback": "example.com",
  },
  "subdomain-authelia": {
    "type": "string",
    "value": "",
    "fallback": "auth",
  },
};

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
  const elements= document.getElementsByClassName(siteVariableName(name));

  if (value === null) {
    console.log(name, "is null");
  }

  const type = variables[name].type;

  if (elements && type) {
    [].slice.call(elements).forEach((element) => {
      element.innerHTML = type === "boolean" ? (value ? variables[name].true : variables[name].false) : value ? value.toString() : "";
    });
  }

  if (name === "domain") {
    siteVariableReplaceDomain(value);
  }
};

const siteVariableReplaceDomain = (value) => {
  if (!value) value = "";

  const relements= document.getElementsByClassName(siteVariableName("domain")+"-regex");

  if (relements) {
    [].slice.call(relements).forEach((element) => {
      element.innerHTML = value.replace(".", "\\.");
    });
  }

  const delements= document.getElementsByClassName(siteVariableName("domain")+"-dn");

  if (delements) {
    [].slice.call(delements).forEach((item) => {
      if (item) {
        item.innerHTML = `DC=${value.replace(".", ",DC=")}`;
      }
    });
  }
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
  let finalValue = fallback;

  // If the browser supports localStorage, setup localStorage elements.
  if (window.localStorage) {
    const type = variables[name].type;
    // If the preference value exists, make sure those tabs are selected.
    const storage = window.localStorage.getItem(siteVariableName(name));

    if (storage) {
      finalValue = type === "boolean" ? storage === "true" : type === "number" ? parseInt(storage) : storage;
      siteVariableReplace(name, finalValue);
    } else {
      siteVariableReplace(name, fallback);
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
  for (const [name, values] of Object.entries(variables)) {
    variables[name].value = siteVariableConfigure(name, values.fallback);
  }

  const save = document.getElementById("site-variables-save");
  if (!save) return;

  const onChangeAutheliaDomain = () => {
    const valueDomain = document.getElementById(siteVariableName("domain")).value.trim();
    const valueSubdomain = document.getElementById(siteVariableName("subdomain-authelia")).value.trim();

    document.getElementById("site-const-authelia-url").value = `https://${valueSubdomain}.${valueDomain}/`;
  };

  const onChangeAutheliaListener = () => {
    const checked = document.getElementById(siteVariableName("tls")).checked;
    const valueHost = document.getElementById(siteVariableName("host")).value.trim();
    const valuePort = document.getElementById(siteVariableName("port")).value.trim();

    document.getElementById("site-const-listen").value = `${checked ? "https" : "http"}://${valueHost}:${valuePort}/`;
  };

  const onSetModalValues = () => {
    for (const [name, values] of Object.entries(variables)) {
      const element = document.getElementById(siteVariableName(name));

      if (element.type === "checkbox") {
        element.checked = values.value;
      } else {
        element.value = values.value;
      }
    }

    onChangeAutheliaDomain();
    onChangeAutheliaListener();
  };

  document.getElementById("site-variables-toggle").addEventListener("click", () => {
    onSetModalValues();
  })

  document.getElementById("site-variables-reset").addEventListener("click", () => {
    for (const [name, values] of Object.entries(variables)) {
      variables[name].value = siteVariableSet(name, values.fallback, variables[name].value);
    }

    onSetModalValues();
  })

  save.addEventListener("click", () => {
    for (const [name, values] of Object.entries(variables)) {
      const element = document.getElementById(siteVariableName(name));
      if (values.type === "boolean") {
        variables[name].value = siteVariableSet(name, element.checked, variables[name].value);
      } else {
        variables[name].value = siteVariableSet(name, element.value.trim(), variables[name].value);
      }
    }
  })

  document.getElementById(siteVariableName("domain")).addEventListener("change", onChangeAutheliaDomain);
  document.getElementById(siteVariableName("domain")).addEventListener("keyup", onChangeAutheliaDomain);
  document.getElementById(siteVariableName("subdomain-authelia")).addEventListener("change", onChangeAutheliaDomain);
  document.getElementById(siteVariableName("subdomain-authelia")).addEventListener("keyup", onChangeAutheliaDomain);
  document.getElementById(siteVariableName("host")).addEventListener("change", onChangeAutheliaListener);
  document.getElementById(siteVariableName("host")).addEventListener("keyup", onChangeAutheliaListener);
  document.getElementById(siteVariableName("port")).addEventListener("change", onChangeAutheliaListener);
  document.getElementById(siteVariableName("port")).addEventListener("keyup", onChangeAutheliaListener);
  document.getElementById(siteVariableName("tls")).addEventListener("change", onChangeAutheliaListener);
};

// Register the 'env' tab group listeners etc. on page load.
customTabsConfigure('env');
customTabsConfigure('session');

siteVariablesConfigure();
