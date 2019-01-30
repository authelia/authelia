import React, { Component, KeyboardEvent, FormEvent } from 'react';
import classnames from 'classnames';

import TextField, { Input } from '@material/react-text-field';
import Button from '@material/react-button';

import styles from '../../assets/scss/components/SecondFactorForm/SecondFactorForm.module.scss';
import CircleLoader, { Status } from '../../components/CircleLoader/CircleLoader';
import Notification from '../Notification/Notification';

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

export type Props = OwnProps & StateProps & DispatchProps;

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
    let u2fStatus = Status.LOADING;
    if (this.props.securityKeyVerified) {
      u2fStatus = Status.SUCCESSFUL;
    } else if (this.props.securityKeyError) {
      u2fStatus = Status.FAILURE;
    }
    return (
      <div className={styles.methodU2f} key='u2f-method'>
        <div className={styles.methodName}>Option {n} - Security Key</div>
        <div>Insert your security key into a USB port and touch the gold disk.</div>
        <div className={styles.imageContainer}>
          <CircleLoader status={u2fStatus}></CircleLoader>
        </div>
        <div className={styles.registerDeviceContainer}>
          <a className={classnames(styles.registerDevice, 'register-u2f')} href="#"
            onClick={this.props.onRegisterSecurityKeyClicked}>
            Register device
          </a>
        </div>
      </div>
    )
  }

  private onOneTimePasswordChanged = (e: FormEvent<HTMLElement>) => {
    this.setState({oneTimePassword: (e.target as HTMLInputElement).value});
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
    return (
      <div className={classnames(styles.methodTotp, 'second-factor-step')} key='totp-method'>
        <div className={styles.methodName}>Option {n} - One-Time Password</div>
        <Notification show={this.props.oneTimePasswordVerificationError !== null}>
          {this.props.oneTimePasswordVerificationError}
        </Notification>
        <TextField
          className={styles.totpField}
          label="One-Time Password"
          outlined={true}>
          <Input
            name='totp-token'
            id='totp-token'
            onChange={this.onOneTimePasswordChanged as any}
            onKeyPress={this.onTotpKeyPressed}
            value={this.state.oneTimePassword} />
        </TextField>
        <div className={styles.registerDeviceContainer}>
          <a className={classnames(styles.registerDevice, 'register-totp')} href="#"
            onClick={this.props.onRegisterOneTimePasswordClicked}>
            Register device
          </a>
        </div>
        <div className={styles.totpButton}>
          <Button
            color="primary"
            raised={true}
            id='totp-button'
            onClick={this.onOneTimePasswordValidationRequested}
            disabled={this.props.oneTimePasswordVerificationInProgress}>
            OK
          </Button>
        </div>
      </div>
    )
  }

  private renderMode() {
    const methods = [];
    let n = 1;
    if (this.props.securityKeySupported) {
      methods.push(this.renderU2f(n));
      n++;
    }
    methods.push(this.renderTotp(n));

    return (
      <div className={styles.methodsContainer}>
        {methods}
      </div>
    );
  }

  render() {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.hello}>Hello <b>{this.props.username}</b></div>
          <div className={styles.logout}>
            <a onClick={this.props.onLogoutClicked} href="#">Logout</a>
          </div>
        </div>
        <div className={styles.body}>
          {this.renderMode()}
        </div>
      </div>
    )
  }
}

export default SecondFactorView;