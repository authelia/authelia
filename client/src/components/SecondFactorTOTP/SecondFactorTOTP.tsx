import React, { FormEvent, KeyboardEvent } from 'react';
import classnames from 'classnames';

import TextField, { Input } from '@material/react-text-field';
import Button from '@material/react-button';
import Notification from '../Notification/Notification';

import styles from '../../assets/scss/components/SecondFactorTOTP/SecondFactorTOTP.module.scss';

export interface OwnProps {
  redirectionUrl: string | null;
}

export interface StateProps {
  oneTimePasswordVerificationInProgress: boolean,
  oneTimePasswordVerificationError: string | null;
}

export interface DispatchProps {
  onRegisterOneTimePasswordClicked: () => void;
  onOneTimePasswordValidationRequested: (token: string) => void;
}

type Props = OwnProps & StateProps & DispatchProps; 

interface State {
  oneTimePassword: string;
}

export default class SecondFactorTOTP extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      oneTimePassword: '',
    }
  }

  private onOneTimePasswordChanged = (e: FormEvent<HTMLInputElement>) => {
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

  render() {
    return (
      <div className={classnames('one-time-password-view')}>
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
          <a className={classnames(styles.registerDevice, 'register-totp')}
            onClick={this.props.onRegisterOneTimePasswordClicked}>
            Register new device
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
}