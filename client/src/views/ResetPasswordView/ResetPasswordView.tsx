import React, { Component, KeyboardEvent, FormEvent } from "react";
import { RouterProps } from "react-router";
import classnames from 'classnames';
import QueryString from 'query-string';

import Button from "@material/react-button";
import TextField, { Input } from "@material/react-text-field";

import styles from '../../assets/scss/views/ResetPasswordView/ResetPasswordView.module.scss';
import Notification from "../../components/Notification/Notification";

export interface StateProps {
  disabled: boolean;
}

export interface DispatchProps {
  onInit: (token: string) => void;
  onPasswordResetRequested: (password: string) => void;
  onCancelClicked: () => void;
}

export type Props = StateProps & DispatchProps & RouterProps;

interface State {
  password1: string;
  password2: string;
  error: string | null,
}

class ResetPasswordView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      password1: '',
      password2: '',
      error: null,
    }
  }

  componentWillMount() {
    if (!this.props.history.location) {
      console.error('There is no location to retrieve query params from...');
      return;
    }
    const params = QueryString.parse(this.props.history.location.search);
    if (!('token' in params)) {
      console.error('Token parameter is expected and not provided');
      return;
    }
    this.props.onInit(params['token'] as string);
  }
  
  private onPasswordResetRequested() {
    if (this.state.password1 && this.state.password1 === this.state.password2) {
      this.props.onPasswordResetRequested(this.state.password1);
    } else {
      this.setState({error: 'The passwords are different.'});
    }
  }

  private onKeyPressed = (e: KeyboardEvent) => {
    if (e.key == 'Enter') {
      this.onPasswordResetRequested();
    }
  }

  private onResetClicked = () => {
    this.onPasswordResetRequested();
  }

  private onPassword1Changed = (e: FormEvent<HTMLElement>) => {
    this.setState({password1: (e.target as HTMLInputElement).value});
  }

  private onPassword2Changed = (e: FormEvent<HTMLElement>) => {
    this.setState({password2: (e.target as HTMLInputElement).value});
  }

  render() {
    return (
      <div>
        <Notification show={this.state.error !== null}>
          {this.state.error}
        </Notification>
        <div>Enter your new password</div>
        <div className={styles.form}>
          <TextField
            className={styles.field}
            outlined={true}
            id="password1"
            label="New password">
            <Input
              type="password"
              key="password1"
              name="password1"
              value={this.state.password1}
              onChange={this.onPassword1Changed}
              disabled={this.props.disabled}/>
          </TextField>
          <TextField
            className={styles.field}
            outlined={true}
            id="password2"
            label="Confirm password">
            <Input
              type="password"
              key="password2"
              name="password2"
              value={this.state.password2}
              onKeyPress={this.onKeyPressed}
              onChange={this.onPassword2Changed}
              disabled={this.props.disabled} />
          </TextField>
          <div className={styles.buttonsContainer}>
            <div className={classnames(styles.buttonContainer, styles.buttonResetContainer)}>
              <Button
                onClick={this.onResetClicked}
                color="primary"
                id="reset-button"
                raised={true}
                disabled={this.props.disabled}
                className={classnames(styles.button, styles.buttonReset)}>
                Reset
              </Button>
            </div>
            <div className={classnames(styles.buttonContainer, styles.buttonCancelContainer)}>
              <Button
                onClick={this.props.onCancelClicked}
                color="primary"
                id="cancel-button"
                raised={true}
                className={classnames(styles.button, styles.buttonCancel)}>
                Cancel
              </Button>
              </div>
          </div>
        </div>
      </div>
    )
  }
}

export default ResetPasswordView;