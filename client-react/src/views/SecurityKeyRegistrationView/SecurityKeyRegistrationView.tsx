import React, { Component } from "react";

import securityKeyImage from '../../assets/images/security-key-hand.png';
import { WithStyles, withStyles } from "@material-ui/core";

import styles from '../../assets/jss/views/SecurityKeyRegistrationView/SecurityKeyRegistrationView';
import { RouteProps } from "react-router";
import QueryString from 'query-string';

interface Props extends WithStyles, RouteProps {
  componentDidMount: (token: string) => void;
}

class SecurityKeyRegistrationView extends Component<Props> {
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
    this.props.componentDidMount(params['token'] as string);
  }

  render() {
    const {classes} = this.props;
    return (
      <div>
        <div className={classes.infoContainer}>Press the gold disk to register your security key</div>
        <div className={classes.imageContainer}>
          <img src={securityKeyImage} alt="security key" />
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(SecurityKeyRegistrationView);