import React, { Component, KeyboardEvent, ChangeEvent } from "react";
import { TextField, Button, WithStyles, withStyles } from "@material-ui/core";
import { RouterProps } from "react-router";
import classnames from 'classnames';
import QueryString from 'query-string';

import styles from '../../assets/jss/views/ResetPasswordView/ResetPasswordView';
import FormNotification from "../../components/FormNotification/FormNotification";

export interface StateProps {
  disabled: boolean;
}

export interface DispatchProps {
  onInit: (token: string) => void;
  onPasswordResetRequested: (password: string) => void;
  onCancelClicked: () => void;
}

export type Props = StateProps & DispatchProps & RouterProps & WithStyles;

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

  private onPassword1Changed = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({password1: e.target.value});
  }

  private onPassword2Changed = (e: ChangeEvent<HTMLInputElement>) => {
    this.setState({password2: e.target.value});
  }

  render() {
    const { classes } = this.props;
    return (
      <div>
        <FormNotification show={this.state.error !== null}>
          {this.state.error}
        </FormNotification>
        <div>Enter your new password</div>
        <div className={classes.form}>
          <TextField
            className={classes.field}
            variant="outlined"
            type="password"
            id="password1"
            value={this.state.password1}
            onChange={this.onPassword1Changed}
            disabled={this.props.disabled}
            label="New password">
          </TextField>
          <TextField
            className={classes.field}
            variant="outlined"
            type="password"
            id="password2"
            value={this.state.password2}
            onKeyPress={this.onKeyPressed}
            onChange={this.onPassword2Changed}
            disabled={this.props.disabled}
            label="Confirm password">
          </TextField>
          <div className={classes.buttonsContainer}>
            <div className={classnames(classes.buttonContainer, classes.buttonResetContainer)}>
              <Button
                onClick={this.onResetClicked}
                variant="contained"
                color="primary"
                disabled={this.props.disabled}
                className={classnames(classes.button, classes.buttonReset)}>
                Reset
              </Button>
            </div>
            <div className={classnames(classes.buttonContainer, classes.buttonCancelContainer)}>
              <Button
                onClick={this.props.onCancelClicked}
                variant="contained"
                color="primary"
                className={classnames(classes.button, classes.buttonCancel)}>
                Cancel
              </Button>
              </div>
          </div>
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(ResetPasswordView);