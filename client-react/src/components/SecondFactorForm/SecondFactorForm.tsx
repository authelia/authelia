import React, { Component, KeyboardEvent, ChangeEvent } from 'react';

import { WithStyles, withStyles, Button, TextField } from '@material-ui/core';

import styles from '../../assets/jss/components/SecondFactorForm/SecondFactorForm';
import CircleLoader, { Status } from '../../components/CircleLoader/CircleLoader';
import FormNotification from '../FormNotification/FormNotification';

export interface OwnProps {
  username: string;
  redirection: string | null;
}

export interface StateProps {
  securityKeySupported: boolean;
  securityKeyVerified: boolean;
  securityKeyError: string | null;

  oneTimePasswordVerificationInProgress: boolean,
  oneTimePasswordVerificationError: string | null;
}

export interface DispatchProps {
  onInit: () => void;
  onLogoutClicked: () => void;
  onRegisterSecurityKeyClicked: () => void;
  onRegisterOneTimePasswordClicked: () => void;

  onOneTimePasswordValidationRequested: (token: string) => void;
}

export type Props = OwnProps & StateProps & DispatchProps & WithStyles;

interface State {
  oneTimePassword: string;
}

class SecondFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      oneTimePassword: '',
    }
  }

  componentWillMount() {
    this.props.onInit();
  }

  private renderU2f(n: number) {
    const { classes } = this.props;
    let u2fStatus = Status.LOADING;
    if (this.props.securityKeyVerified) {
      u2fStatus = Status.SUCCESSFUL;
    } else if (this.props.securityKeyError) {
      u2fStatus = Status.FAILURE;
    }
    return (
      <div className={classes.methodU2f} key='u2f-method'>
        <div className={classes.methodName}>Option {n} - Security Key</div>
        <div>Insert your security key into a USB port and touch the gold disk.</div>
        <div className={classes.imageContainer}>
          <CircleLoader status={u2fStatus}></CircleLoader>
        </div>
        <div className={classes.registerDeviceContainer}>
          <a className={classes.registerDevice} href="#"
            onClick={this.props.onRegisterSecurityKeyClicked}>
            Register device
          </a>
        </div>
      </div>
    )
  }

  private onOneTimePasswordChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({oneTimePassword: e.target.value});
  }

  private onTotpKeyPressed = (e: KeyboardEvent) => {
    if (e.key === 'Enter') {
      this.onOneTimePasswordValidationRequested();
    }
  }

  private onOneTimePasswordValidationRequested = () => {
    if (this.props.oneTimePasswordVerificationInProgress) return;
    this.props.onOneTimePasswordValidationRequested(this.state.oneTimePassword);
  }

  private renderTotp(n: number) {
    const { classes } = this.props;
    return (
      <div className={classes.methodTotp} key='totp-method'>
        <div className={classes.methodName}>Option {n} - One-Time Password</div>
        <FormNotification show={this.props.oneTimePasswordVerificationError !== null}>
          {this.props.oneTimePasswordVerificationError}
        </FormNotification>
        <TextField
          className={classes.totpField}
          name="totp-token"
          id="totp-token"
          variant="outlined"
          label="One-Time Password"
          onChange={this.onOneTimePasswordChanged}
          onKeyPress={this.onTotpKeyPressed}>
        </TextField>
        <div className={classes.registerDeviceContainer}>
          <a className={classes.registerDevice} href="#"
            onClick={this.props.onRegisterOneTimePasswordClicked}>
            Register device
          </a>
        </div>
        <Button
          className={classes.totpButton}
          variant="contained"
          color="primary"
          onClick={this.onOneTimePasswordValidationRequested}
          disabled={this.props.oneTimePasswordVerificationInProgress}>
          OK
        </Button>
      </div>
    )
  }

  private renderMode() {
    const { classes } = this.props;
    const methods = [];
    let n = 1;
    if (this.props.securityKeySupported) {
      methods.push(this.renderU2f(n));
      n++;
    }
    methods.push(this.renderTotp(n));

    return (
      <div className={classes.methodsContainer}>
        {methods}
      </div>
    );
  }

  render() {
    const { classes } = this.props;
    return (
      <div className={classes.container}>
        <div className={classes.header}>
          <div className={classes.hello}>Hello <b>{this.props.username}</b></div>
          <div className={classes.logout}>
            <a onClick={this.props.onLogoutClicked} href="#">Logout</a>
          </div>
        </div>
        <div className={classes.body}>
          {this.renderMode()}
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(SecondFactorView);