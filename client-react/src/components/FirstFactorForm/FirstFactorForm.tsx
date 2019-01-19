import React, { Component, KeyboardEvent, ChangeEvent } from "react";

import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import { Link } from "react-router-dom";
import { WithStyles, withStyles } from "@material-ui/core";

import styles from '../../assets/jss/components/FirstFactorForm/FirstFactorForm';
import FormNotification from "../../components/FormNotification/FormNotification";

import CheckBoxOutlineBlankIcon from '@material-ui/icons/CheckBoxOutlineBlank'
import CheckBoxIcon from '@material-ui/icons/CheckBox';

export interface StateProps {
  formDisabled: boolean;
  error: string | null;
}

export interface DispatchProps {
  onAuthenticationRequested(username: string, password: string): void;
}

export type Props = StateProps & DispatchProps & WithStyles;

interface State {
  username: string;
  password: string;
  rememberMe: boolean;
}

class FirstFactorForm extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      username: '',
      password: '',
      rememberMe: false,
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
        <FormNotification
          show={this.props.error != null}>
          {this.props.error || ''}
        </FormNotification>
        <div className={classes.fields}>
          <div className={classes.field}>
            <TextField
              className={classes.input}
              variant="outlined"
              id="username"
              label="Username"
              disabled={this.props.formDisabled}
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
              disabled={this.props.formDisabled}
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
              disabled={this.props.formDisabled}>
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

  private authenticate() {
    this.props.onAuthenticationRequested(
      this.state.username,
      this.state.password);
  }
}

export default withStyles(styles)(FirstFactorForm);