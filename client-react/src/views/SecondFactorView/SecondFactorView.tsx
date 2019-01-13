import React, { Component } from 'react';

import { WithStyles, withStyles, Button, TextField } from '@material-ui/core';

import styles from '../../assets/jss/views/SecondFactorView/SecondFactorView';
import securityKeyImage from '../../assets/images/security-key-hand.png';
import StateSynchronizer from '../../containers/components/StateSynchronizer/StateSynchronizer';
import { RouterProps, Redirect } from 'react-router';
import RemoteState from '../../reducers/Portal/RemoteState';
import AuthenticationLevel from '../../types/AuthenticationLevel';
import { WithState } from '../../components/StateSynchronizer/WithState';

type Mode = 'u2f' | 'totp';

export interface Props extends WithStyles, RouterProps, WithState {
  onLogoutClicked: () => void;
  onRegisterSecurityKeyClicked: () => void;
  onRegisterOneTimePasswordClicked: () => void;
  onStateLoaded: (state: RemoteState) => void;
};

interface State {
  mode: Mode;
  remoteState: RemoteState | null;
}

class SecondFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      mode: 'u2f',
      remoteState: null,
    }
  }

  private toggleMode = () => {
    if (this.state.mode === 'u2f') {
      this.setState({mode: 'totp'});
    } else if (this.state.mode === 'totp') {
      this.setState({mode: 'u2f'});
    }
  }

  private renderU2f() {
    const { classes } = this.props;
    return (
      <div>
        <div className={classes.imageContainer}>
          <img src={securityKeyImage} alt='security key' className={classes.image}/>
        </div>
        <div>Insert your security key into a USB port and touch the gold disk.</div>
      </div>
    )
  }

  private renderTotp() {
    const { classes } = this.props;
    return (
      <div>
        <div>Provide a one-time password.</div>
        <TextField
          className={classes.totpField}
          name="password"
          id="password"
          variant="outlined"
          label="Password">
        </TextField>
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
    if (this.state.mode === 'u2f') {
      return this.renderU2f();
    } else if (this.state.mode === 'totp') {
      return this.renderTotp();
    }
  }

  private onRegisterClicked = () => {
    const mode = this.state.mode;
    if (mode === 'u2f') {
      this.props.onRegisterSecurityKeyClicked();
    } else {
      this.props.onRegisterOneTimePasswordClicked();
    }
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
        <div className={classes.footer}>
          <a 
            className={classes.otherMethod}
            href="#"
            onClick={this.toggleMode}>
            {this.state.mode === 'u2f' ? 'Use one-time password' : 'Use security key'}
          </a>
          <a className={classes.registerDevice} href="#" onClick={this.onRegisterClicked}>
            Register device
          </a>
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