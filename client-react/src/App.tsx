import React, { Component } from 'react';
import './App.css';

import { Router, Route, Switch } from "react-router-dom";
import { routes } from './routes/index';
import { createBrowserHistory } from 'history';
import { createStore, applyMiddleware } from 'redux';
import reducer from './reducers';
import { Provider } from 'react-redux';
import thunk from 'redux-thunk';

const history = createBrowserHistory();
const store = createStore(
  reducer,
  applyMiddleware(thunk)
);

class App extends Component {
  render() {
    return (
      <Provider store={store}>
        <Router history={history}>
          <div className="App">
            <Switch>
              {routes.map((r, key) => {
                return <Route path={r.path} component={r.component} key={key}/>
              })}
            </Switch>
          </div>
        </Router>
      </Provider>
    );
  }
}

export default App;
