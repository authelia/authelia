import React, { Component } from 'react';
import './App.scss';

import { Route, Switch } from "react-router-dom";
import { routes } from './routes/index';
import { createHashHistory } from 'history';
import { createStore, applyMiddleware, compose } from 'redux';
import reducer from './reducers';
import { Provider } from 'react-redux';
import thunk from 'redux-thunk';
import { routerMiddleware, ConnectedRouter } from 'connected-react-router';

const history = createHashHistory();
const store = createStore(
  reducer(history),
  compose(
    applyMiddleware(
      routerMiddleware(history),
      thunk
    )
  )
);

class App extends Component {
  render() {
    return (
      <Provider store={store}>
        <ConnectedRouter history={history}>
          <div className="App">
            <Switch>
              {routes.map((r, key) => {
                return <Route path={r.path} component={r.component} key={key}/>
              })}
            </Switch>
          </div>
        </ConnectedRouter>
      </Provider>
    );
  }
}

export default App;
