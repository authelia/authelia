import React, { Component, KeyboardEvent, ChangeEvent } from "react";

import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import { Link } from "react-router-dom";
import { RouterProps, RouteProps } from "react-router";
import { WithStyles, withStyles } from "@material-ui/core";

import firstFactorViewStyles from '../../assets/jss/views/FirstFactorView/FirstFactorView';
import FormNotification from "../../components/FormNotification/FormNotification";

import CheckBoxOutlineBlankIcon from '@material-ui/icons/CheckBoxOutlineBlank'
import CheckBoxIcon from '@material-ui/icons/CheckBox';
import StateSynchronizer from "../../containers/components/StateSynchronizer/StateSynchronizer";
import RemoteState from "../../reducers/Portal/RemoteState";

export interface Props extends RouteProps, RouterProps, WithStyles {
  onAuthenticationRequested(username: string, password: string): void;
}

interface State {
  rememberMe: boolean;
  username: string;
  password: string;
  loginButtonDisabled: boolean;
  errorMessage: string | null;
  remoteState: RemoteState | null;
}

class FirstFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      rememberMe: false,
      username: '',
      password: '',
      loginButtonDisabled: false,
      errorMessage: null,
      remoteState: null,
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

  private renderWithState() {
    const { classes } = this.props;
    return (
      <div>
        <FormNotification
          show={this.state.errorMessage != null}>
          {this.state.errorMessage || ''}
        </FormNotification>
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
                    icon={<CheckBoxOutlineBlankIcon fontSize="small" />}
                    checkedIcon={<CheckBoxIcon fontSize="small" />}
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

  render() {
    return (
      <div>
        <StateSynchronizer
          onLoaded={(remoteState) => this.setState({remoteState})}/>
        {this.state.remoteState ? this.renderWithState() : null}
      </div>
    )
  }

  private authenticate() {
    this.setState({loginButtonDisabled: true});
    this.props.onAuthenticationRequested(
      this.state.username,
      this.state.password);
    this.setState({errorMessage: null});
  }

  onFailure = (error: string) => {
    this.setState({
      loginButtonDisabled: false,
      errorMessage: 'An error occured. Your username/password are probably wrong.'
    });
  }
}

export default withStyles(firstFactorViewStyles)(FirstFactorView);