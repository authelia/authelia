import React, { Component, KeyboardEvent, FormEvent } from "react";

import TextField, {Input} from '@material/react-text-field';
import Button from '@material/react-button';
import Checkbox from '@material/react-checkbox';

import { Link } from "react-router-dom";

import Notification from "../../components/Notification/Notification";

import styles from '../../assets/scss/components/FirstFactorForm/FirstFactorForm.module.scss';

export interface OwnProps {
  redirectionUrl: string | null;
}

export interface StateProps {
  formDisabled: boolean;
  error: string | null;
}

export interface DispatchProps {
  onAuthenticationRequested(username: string, password: string, rememberMe: boolean): Promise<void>;
}

export type Props = OwnProps & StateProps & DispatchProps;

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

  onUsernameChanged = (e: FormEvent<HTMLElement>) => {
    const val = (e.target as HTMLInputElement).value;
    this.setState({username: val});
  }

  onPasswordChanged = (e: FormEvent<HTMLElement>) => {
    const val = (e.target as HTMLInputElement).value;
    this.setState({password: val});
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
    return (
      <div className='first-factor-step'>
        <Notification
          show={this.props.error != null}
          className={styles.notification}>
          {this.props.error || ''}
        </Notification>
        <div className={styles.fields}>
          <div className={styles.field}>
            <TextField
              className={styles.input}
              label="Username"
              outlined={true}>
              <Input
                id="username"
                onChange={this.onUsernameChanged}
                disabled={this.props.formDisabled}
                value={this.state.username}/>
            </TextField>
          </div>
          <div className={styles.field}>
            <TextField
              className={styles.input}
              label="Password"
              outlined={true}>
              <Input
                id="password"
                type="password"
                disabled={this.props.formDisabled}
                onChange={this.onPasswordChanged}
                onKeyPress={this.onPasswordKeyPressed}
                value={this.state.password} />
            </TextField>
          </div>
        </div>
        <div>
          <div className={styles.buttons}>
            <Button
              onClick={this.onLoginClicked}
              color='primary'
              raised={true}
              id='login-button'
              disabled={this.props.formDisabled}>
              Login
            </Button>
          </div>
          <div className={styles.controls}>
            <div className={styles.rememberMe}>
              <Checkbox
                nativeControlId='remember-checkbox'
                checked={this.state.rememberMe}
                onChange={this.toggleRememberMe}
              />
              <label htmlFor='remember-checkbox'>Remember me</label>
            </div>
            <div className={styles.resetPassword}>
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
      this.state.password,
      this.state.rememberMe)
      .catch((err: Error) => console.error(err))
      .finally(() => {
        this.setState({username: '', password: ''});
      })
  }
}

export default FirstFactorForm;