import React, { Component, ChangeEvent, KeyboardEvent } from "react";
import { TextField, WithStyles, withStyles, Button } from "@material-ui/core";
import classnames from 'classnames';

import styles from '../../assets/jss/views/ForgotPasswordView/ForgotPasswordView';

export interface StateProps {
  disabled: boolean;
}

export interface DispatchProps {
  onPasswordResetRequested: (username: string) => void;
  onCancelClicked: () => void;
}

export type Props = StateProps & DispatchProps & WithStyles;

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

  private onUsernameChanged = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({username: e.target.value});
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
    const { classes } = this.props;
    return (
      <div>
        <div>What's your username?</div>
        <div className={classes.form}>
          <TextField
            className={classes.field}
            variant="outlined"
            id="username"
            label="Username"
            onChange={this.onUsernameChanged}
            onKeyPress={this.onKeyPressed}
            value={this.state.username}
            disabled={this.props.disabled}>
          </TextField>
          <div className={classes.buttonsContainer}>
            <div className={classnames(classes.buttonContainer, classes.buttonConfirmContainer)}>
              <Button
                onClick={this.onPasswordResetRequested}
                variant="contained"
                color="primary"
                className={classes.buttonConfirm}
                disabled={this.props.disabled}>
                Next
              </Button>
            </div>
            <div className={classnames(classes.buttonContainer, classes.buttonCancelContainer)}>
              <Button
                onClick={this.props.onCancelClicked}
                variant="contained"
                color="primary"
                className={classes.buttonCancel}>
                Cancel
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(ForgotPasswordView);