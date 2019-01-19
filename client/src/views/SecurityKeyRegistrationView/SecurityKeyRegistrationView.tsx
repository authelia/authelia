import React, { Component } from "react";

import { WithStyles, withStyles, Button } from "@material-ui/core";

import styles from '../../assets/jss/views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RouteProps, RouterProps } from "react-router";
import QueryString from 'query-string';
import FormNotification from "../../components/FormNotification/FormNotification";
import CircleLoader, { Status } from "../../components/CircleLoader/CircleLoader";

export interface Props extends WithStyles, RouteProps, RouterProps {
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
    const { classes } = this.props;
    return (
      <div>
        <FormNotification show={true}>
          {this.props.error}
        </FormNotification>
        <div className={classes.retryButtonContainer}>
          <Button
            variant="contained"
            color="primary"
            onClick={this.props.onBackClicked}>
            Back
          </Button>
        </div>
      </div>
    )
  }

  private renderRegistering() {
    const { classes } = this.props;
    let status = Status.LOADING;
    if (this.props.deviceRegistered === true) {
      status = Status.SUCCESSFUL;
    } else if (this.props.error) {
      status = Status.FAILURE;
    }
    return (
      <div>
        <div className={classes.infoContainer}>Press the gold disk to register your security key</div>
        <div className={classes.imageContainer}>
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

export default withStyles(styles)(SecurityKeyRegistrationView);