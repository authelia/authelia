import React, { Component } from "react";

import Button from "@material/react-button";

import styles from '../../assets/scss/views/SecurityKeyRegistrationView/SecurityKeyRegistrationView.module.scss';
import { RouteProps, RouterProps } from "react-router";
import QueryString from 'query-string';
import Notification from "../../components/Notification/Notification";
import CircleLoader, { Status } from "../../components/CircleLoader/CircleLoader";

export interface Props extends RouteProps, RouterProps {
  deviceRegistered: boolean | null;
  error: string | null;
  onInit: (token: string) => void;
  onBackClicked: () => void;
}

class SecurityKeyRegistrationView extends Component<Props> {
  componentDidMount() {
    if (this.props.deviceRegistered) return;

    if (!this.props.location) {
      console.error('There is no location to retrieve query params from...');
      return;
    }
    const params = QueryString.parse(this.props.location.search);
    if (!('token' in params)) {
      console.error('Token parameter is expected and not provided');
      return;
    }
    this.props.onInit(params['token'] as string);
  }

  private renderError() {
    return (
      <div>
        <Notification show={true}>
          {this.props.error}
        </Notification>
        <div className={styles.retryButtonContainer}>
          <Button
            color="primary"
            raised={true}
            onClick={this.props.onBackClicked}>
            Back
          </Button>
        </div>
      </div>
    )
  }

  private renderRegistering() {
    let status = Status.LOADING;
    if (this.props.deviceRegistered === true) {
      status = Status.SUCCESSFUL;
    } else if (this.props.error) {
      status = Status.FAILURE;
    }
    return (
      <div>
        <div className={styles.infoContainer}>Press the gold disk to register your security key</div>
        <div className={styles.imageContainer}>
          <CircleLoader status={status}></CircleLoader>
        </div>
      </div>
    )
  }

  render() {
    if (this.props.error) {
      return this.renderError();
    }

    return this.renderRegistering();
  }
}

export default SecurityKeyRegistrationView;