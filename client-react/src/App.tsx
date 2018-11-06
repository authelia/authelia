import React, { Component } from 'react';
import './App.css';

import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import { FirstFactor } from './first-factor';
import { SecondFactor } from './second-factor';

class App extends Component {
  render() {
    return (
      <Router>
        <div className="App">
          <Route exact path="/" component={FirstFactor} />
          <Route exact path="/2fa" component={SecondFactor} />
        </div>
      </Router>
    );
  }
}

export default App;
