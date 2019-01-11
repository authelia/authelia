import React, { Component, KeyboardEvent, ChangeEvent } from "react";

import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import { Link } from "react-router-dom";
import { RouterProps } from "react-router";
import { WithStyles, withStyles } from "@material-ui/core";

import firstFactorViewStyles from '../../assets/jss/views/FirstFactorView/FirstFactorView';

interface Props extends RouterProps, WithStyles {}

interface State {
  rememberMe: boolean;
  username: string;
  password: string;
  loginButtonDisabled: boolean;
}

class FirstFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      rememberMe: false,
      username: '',
      password: '',
      loginButtonDisabled: false,
    }
  }

  toggleRememberMe = () => {
    this.setState({
      rememberMe: !(this.state.rememberMe)
    })
  }

  onUsernameChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({username: e.target.value});
  }

  onPasswordChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({password: e.target.value});
  }

  onLoginClicked = () => {
    this.authenticate();
  }

  onPasswordKeyPressed = (e: KeyboardEvent) => {
    if (e.key === 'Enter') {
      this.authenticate();
    }
  }

  render() {
    const { classes } = this.props;
    return (
      <div>
        <div className={classes.fields}>
          <div className={classes.field}>
            <TextField
              className={classes.input}
              variant="outlined"
              id="username"
              label="Username"
              onChange={this.onUsernameChanged}>
            </TextField>
          </div>
          <div className={classes.field}>
            <TextField
              className={classes.input}
              id="password"
              variant="outlined"
              label="Password"
              type="password"
              onChange={this.onPasswordChanged}
              onKeyPress={this.onPasswordKeyPressed}>
            </TextField>
          </div>
        </div>
        <div>
          <div className={classes.buttons}>
            <Button
              onClick={this.onLoginClicked}
              variant="contained"
              color="primary"
              disabled={this.state.loginButtonDisabled}>
              Login
            </Button>
          </div>
          <div className={classes.controls}>
            <div className={classes.rememberMe}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={this.state.rememberMe}
                    onChange={this.toggleRememberMe}
                    color="primary"
                  />
                }
                label="Remember me"
              />
            </div>
            <div className={classes.resetPassword}>
              <Link to="/forgot-password">Forgot password?</Link>
            </div>
          </div>
        </div>
      </div>
    )
  }

  private authenticate() {
    this.setState({loginButtonDisabled: true})
    fetch('/api/firstfactor', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: this.state.username,
        password: this.state.password,
      })
    }).then(async (res) => {
      const json = await res.json();
      if ('error' in json) {
        console.log('ERROR!');
        this.setState({loginButtonDisabled: false});
        return;
      }
      this.props.history.push('/2fa');
    });
  }
}

export default withStyles(firstFactorViewStyles)(FirstFactorView);