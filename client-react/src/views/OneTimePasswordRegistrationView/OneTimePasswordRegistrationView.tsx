import React, { Component } from "react";

import { WithStyles, withStyles, TextField } from "@material-ui/core";

import styles from '../../assets/jss/views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView';
import { RouteProps } from "react-router";
import QueryString from 'query-string';

import QRCode from 'qrcode.react';

export type OnSuccess = (secret: Secret) => void;
export type OnFailure = (err: Error) => void;

export interface Props extends WithStyles, RouteProps {
  componentDidMount: (token: string, onSuccess: OnSuccess, onFailure: OnFailure) => void;
}

export interface Secret {
  otp_url: string;
  base32_secret: string;
}

interface State {
  secret: Secret | null;
}

class OneTimePasswordRegistrationView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      secret: {
        otp_url: 'https://coucou',
        base32_secret: 'coucou',
      },
    }
  }
  componentDidMount() {
    if (!this.props.location) {
      console.error('There is no location to retrieve query params from...');
      return;
    }
    const params = QueryString.parse(this.props.location.search);
    if (!('token' in params)) {
      console.error('Token parameter is expected and not provided');
      return;
    }
    this.props.componentDidMount(
      params['token'] as string,
      this.onSuccess,
      this.onFailure);
  }

  onSuccess = (secret: Secret) => {
    this.setState({secret});
  }

  onFailure = (err: Error) => {}

  private renderWithSecret(secret: Secret) {
    const { classes } = this.props;
    return (
      <div>
        <div className={classes.secretContainer}>
          <TextField
            id="totp-secret"
            label="Key"
            defaultValue={secret.base32_secret}
            className={classes.textField}
            margin="normal"
            InputProps={{
              readOnly: true,
            }}
            variant="outlined"
          />
        </div>
        <div className={classes.qrcodeContainer}>
          {this.state.secret ? <QRCode value={this.state.secret.otp_url}></QRCode> : null}
        </div>
      </div>
    )
  }

  render() {
    return this.state.secret ? this.renderWithSecret(this.state.secret) : null;
  }
}

export default withStyles(styles)(OneTimePasswordRegistrationView);