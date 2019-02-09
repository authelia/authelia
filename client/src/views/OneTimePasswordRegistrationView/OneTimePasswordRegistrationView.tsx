import React, { Component } from "react";
import classnames from 'classnames';

import Button from "@material/react-button";

import styles from '../../assets/scss/views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView.module.scss';
import { RouteProps, RouterProps } from "react-router";
import QueryString from 'query-string';

import QRCode from 'qrcode.react';
import Notification from "../../components/Notification/Notification";

import googleStoreImage from '../../assets/images/googleplay-badge.svg';
import appleStoreImage from '../../assets/images/applestore-badge.svg';
import { Secret } from "./Secret";
import CircleLoader, { Status } from "../../components/CircleLoader/CircleLoader";

export interface Props extends RouteProps, RouterProps {
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
    return (
      <div>
        <div className={styles.text}>
          Register your device by scanning the barcode or adding the key.
        </div>
        <div className={styles.secretContainer}>
          <div className={classnames(styles.qrcodeContainer, 'qrcode')}>
            <QRCode value={secret.otpauth_url} size={180} level="Q"></QRCode>
          </div>
          <div className={classnames(styles.otpauthContainer, 'otpauth-secret')}>{secret.otpauth_url}</div>
          <div className={classnames(styles.base32Container, 'base32-secret')}>{secret.base32_secret}</div>
        </div>
        <div className={styles.loginButtonContainer}>
          <Button
            color="primary"
            raised={true}
            onClick={this.props.onLoginClicked}>
            Login
          </Button>
        </div>
        <div className={styles.needGoogleAuthenticator}>
          <div className={styles.needGoogleAuthenticatorText}>Need Google Authenticator?</div>
          <img src={appleStoreImage} className={styles.store} alt='Google Authenticator on Apple Store'/>
          <img src={googleStoreImage} className={styles.store} alt='Google Authenticator on Google Store'/>
        </div>
      </div>
    )
  }

  private renderError() {
    return (
      <div>
        <Notification show={true}>
          <div>{this.props.error}</div>
        </Notification>
        <div className={styles.buttonContainer}>
          <Button
            color="primary"
            raised={true}
            className={styles.button}
            onClick={this.props.onRetryClicked}>
            Retry
          </Button>
          <Button
            color="primary"
            raised={true}
            className={styles.button}
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
    return (
      <div>
        <div>One-Time password secret is being generated...</div>
        <div className={styles.progressContainer}><CircleLoader status={Status.LOADING} /></div>
      </div>
    )
  }

  render() {
    return !this.props.secret && !this.props.error
      ? this.renderLoading()
      : this.renderSecret();
  }
}

export default OneTimePasswordRegistrationView;