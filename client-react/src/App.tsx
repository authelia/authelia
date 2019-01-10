import React, { Component } from 'react';
import './App.css';

import { Router, Route, Switch } from "react-router-dom";
import { routes } from './routes/index';
import { createBrowserHistory } from 'history';

const history = createBrowserHistory();

class App extends Component {
  render() {
    return (
      <Router history={history}>
        <div className="App">
          <Switch>
            {routes.map((r, key) => {
              return <Route path={r.path} component={r.component} key={key}/>
            })}
          </Switch>
        </div>
      </Router>
    );
  }
}

export default App;
