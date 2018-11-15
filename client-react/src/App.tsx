import React, { Component } from 'react';
import './App.css';

import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import { FirstFactor } from './pages/first-factor/first-factor';
import { SecondFactor } from './pages/second-factor/second-factor';
import ConfirmationSent from './pages/confirmation-sent/confirmation-sent';

class App extends Component {
  render() {
    return (
      <Router>
        <div className="App">
          <Route exact path="/" component={FirstFactor} />
          <Route exact path="/2fa" component={SecondFactor} />
          <Route exact path="/confirmation" component={ConfirmationSent} />
        </div>
      </Router>
    );
  }
}

export default App;
