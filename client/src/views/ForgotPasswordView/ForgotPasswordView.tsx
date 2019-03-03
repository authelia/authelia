import React, { Component, KeyboardEvent, FormEvent } from "react";
import classnames from 'classnames';
import Button from "@material/react-button";
import TextField, { Input } from "@material/react-text-field";

import styles from '../../assets/scss/views/ForgotPasswordView/ForgotPasswordView.module.scss';

export interface StateProps {
  disabled: boolean;
}

export interface DispatchProps {
  onPasswordResetRequested: (username: string) => void;
  onCancelClicked: () => void;
}

export type Props = StateProps & DispatchProps;

interface State {
  username: string;
}

class ForgotPasswordView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      username: '',
    }
  }

  private onUsernameChanged = (e: FormEvent<HTMLElement>) => {
    this.setState({username: (e.target as HTMLInputElement).value});
  }

  private onKeyPressed = (e: KeyboardEvent) => {
    if (e.key == 'Enter') {
      this.onPasswordResetRequested();
    }
  }

  private onPasswordResetRequested = () => {
    if (this.state.username.length == 0) return;
    this.props.onPasswordResetRequested(this.state.username);
  }

  render() {
    return (
      <div>
        <div>What's your username?</div>
        <div className={styles.form}>
          <TextField
            className={styles.field}
            outlined={true}
            label="Username">
            <Input
              id="username"
              name="username"
              onChange={this.onUsernameChanged}
              onKeyPress={this.onKeyPressed}
              value={this.state.username}
              disabled={this.props.disabled} />
          </TextField>
          <div className={styles.buttonsContainer}>
            <div className={classnames(styles.buttonContainer, styles.buttonConfirmContainer)}>
              <Button
                onClick={this.onPasswordResetRequested}
                color="primary"
                id="next-button"
                raised={true}
                className={styles.buttonConfirm}
                disabled={this.props.disabled}>
                Next
              </Button>
            </div>
            <div className={classnames(styles.buttonContainer, styles.buttonCancelContainer)}>
              <Button
                onClick={this.props.onCancelClicked}
                color="primary"
                raised={true}
                id="cancel-button"
                className={styles.buttonCancel}>
                Cancel
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default ForgotPasswordView;