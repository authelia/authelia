import './hooks/AssetPath';
import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import * as serviceWorker from './serviceWorker';
import { ThemeProvider, CssBaseline } from '@material-ui/core';
import { useTheme } from './hooks/Theme';
import * as themes from './themes';

function Theme() {
  switch (useTheme()) {
    case 'dark':
      return themes.dark;
    case 'light':
      return themes.light;
    case 'custom':
      return themes.custom;
    default:
      return themes.light;
  }
}

ReactDOM.render(
  <ThemeProvider theme={Theme()}>
    <CssBaseline />
    <App />
  </ThemeProvider>
  , document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
