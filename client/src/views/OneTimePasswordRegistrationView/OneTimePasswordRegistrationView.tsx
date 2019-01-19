import React, { Component } from "react";

import { WithStyles, withStyles, CircularProgress, Button } from "@material-ui/core";

import styles from '../../assets/jss/views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView';
import { RouteProps, RouterProps } from "react-router";
import QueryString from 'query-string';

import QRCode from 'qrcode.react';
import FormNotification from "../../components/FormNotification/FormNotification";

import googleStoreImage from '../../assets/images/googleplay-badge.svg';
import appleStoreImage from '../../assets/images/applestore-badge.svg';
import { Secret } from "./Secret";

export interface Props extends WithStyles, RouteProps, RouterProps {
  secret: Secret | null;
  error: string | null;
  onInit: (token: string) => void;
  onRetryClicked: () => void;
  onCancelClicked: () => void;
  onLoginClicked: () => void;
}

class OneTimePasswordRegistrationView extends Component<Props> {
  private token: string | null;
  constructor(props: Props) {
    super(props);
    this.token = null;
  }

  componentWillMount() {
    // If secret is already populated, we skip onInit (for testing purposes).
    if (this.props.secret) return;

    if (!this.props.location) {
      console.error('There is no location to retrieve query params from...');
      return;
    }
    const params = QueryString.parse(this.props.location.search);
    if (!('token' in params)) {
      console.error('Token parameter is expected and not provided');
      return;
    }
    this.token = params['token'] as string;
    this.props.onInit(this.token);
  }

  private renderWithSecret(secret: Secret) {
    const { classes } = this.props;
    return (
      <div>
        <div className={classes.text}>
          Register your device by scanning the barcode or adding the key.
        </div>
        <div className={classes.secretContainer}>
          <div className={classes.qrcodeContainer}>
            <QRCode value={secret.otpauth_url} size={180} level="Q"></QRCode>
          </div>
          <div className={classes.base32Container}>{secret.base32_secret}</div>
        </div>
        <div className={classes.loginButtonContainer}>
          <Button
            color="primary"
            variant="contained"
            onClick={this.props.onLoginClicked}>
            Login
          </Button>
        </div>
        <div className={classes.needGoogleAuthenticator}>
          <div className={classes.needGoogleAuthenticatorText}>Need Google Authenticator?</div>
          <img src={appleStoreImage} className={classes.store} alt='Google Authenticator on Apple Store'/>
          <img src={googleStoreImage} className={classes.store} alt='Google Authenticator on Google Store'/>
        </div>
      </div>
    )
  }

  private renderError() {
    const {classes} = this.props;
    return (
      <div>
        <FormNotification show={true}>
          <div>{this.props.error}</div>
        </FormNotification>
        <div className={classes.buttonContainer}>
          <Button
            variant="contained"
            color="primary"
            className={classes.button}
            onClick={this.props.onRetryClicked}>
            Retry
          </Button>
          <Button
            variant="contained"
            color="primary"
            className={classes.button}
            onClick={this.props.onCancelClicked}>
            Cancel
          </Button>
        </div>
      </div>
    );
  }

  private renderSecret() {
    return this.props.secret
      ? this.renderWithSecret(this.props.secret)
      : this.renderError();
  }

  private renderLoading() {
    const { classes } = this.props;
    return (
      <div>
        <div>One-Time password secret is being generated...</div>
        <div className={classes.progressContainer}><CircularProgress /></div>
      </div>
    )
  }

  render() {
    return !this.props.secret && !this.props.error
      ? this.renderLoading()
      : this.renderSecret();
  }
}

export default withStyles(styles)(OneTimePasswordRegistrationView);