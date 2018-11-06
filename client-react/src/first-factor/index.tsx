import React, { Component } from "react";

import TextField from '@material-ui/core/TextField';

export class FirstFactor extends Component {
  render() {
    return (
      <div className="frame">
        <TextField
          id="username"
          label="Username">
        </TextField>
        <TextField
          id="password"
          label="Password"
          type="password">
        </TextField>
      </div>
    )
  }
}