import React, { Component } from 'react';

import { WithStyles, withStyles, Button, TextField } from '@material-ui/core';

import styles from '../../assets/jss/views/SecondFactorView/SecondFactorView';
import StateSynchronizer from '../../containers/components/StateSynchronizer/StateSynchronizer';
import { RouterProps, Redirect } from 'react-router';
import RemoteState from '../../reducers/Portal/RemoteState';
import AuthenticationLevel from '../../types/AuthenticationLevel';
import { WithState } from '../../components/StateSynchronizer/WithState';
import CircleLoader, { Status } from '../../components/CircleLoader/CircleLoader';

export interface Props extends WithStyles, RouterProps, WithState {
  securityKeySupported: boolean;
  securityKeyVerified: boolean;
  securityKeyError: string | null;

  onLogoutClicked: () => void;
  onRegisterSecurityKeyClicked: () => void;
  onRegisterOneTimePasswordClicked: () => void;
  onStateLoaded: (state: RemoteState) => void;
};

interface State {
  remoteState: RemoteState | null;
}

class SecondFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      remoteState: null,
    }
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

  private renderTotp(n: number) {
    const { classes } = this.props;
    return (
      <div className={classes.methodTotp} key='totp-method'>
        <div className={classes.methodName}>Option {n} - One-Time Password</div>
        <TextField
          className={classes.totpField}
          name="password"
          id="password"
          variant="outlined"
          label="One-Time Password">
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
          color="primary">
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

  private renderWithState(state: RemoteState) {
    if (state.authentication_level < AuthenticationLevel.ONE_FACTOR) {
      return <Redirect to='/' key='redirect' />;
    }

    const { classes } = this.props;
    return (
      <div className={classes.container}>
        <div className={classes.header}>
          <div className={classes.hello}>Hello <b>{state.username}</b></div>
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

  onStateLoaded = (remoteState: RemoteState) => {
    this.setState({remoteState});
    this.props.onStateLoaded(remoteState);
  }

  render() {
    return (
      <div>
        <StateSynchronizer
          onLoaded={this.onStateLoaded}/>
        {this.state.remoteState ? this.renderWithState(this.state.remoteState) : null}
      </div>
    )
  }
}

export default withStyles(styles)(SecondFactorView);