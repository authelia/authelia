import React, { Component } from 'react';
import './App.css';

import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import { FirstFactor } from './pages/first-factor/first-factor';
import { SecondFactor } from './pages/second-factor/second-factor';

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
