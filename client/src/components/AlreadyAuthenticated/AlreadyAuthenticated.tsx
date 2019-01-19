import React, { Component } from "react";

import styles from '../../assets/jss/components/AlreadyAuthenticated/AlreadyAuthenticated';
import { WithStyles, withStyles, Button } from "@material-ui/core";
import CircleLoader, { Status } from "../CircleLoader/CircleLoader";

export interface OwnProps {
  username: string;
}

export interface DispatchProps {
  onLogoutClicked: () => void;
}

export type Props = OwnProps & DispatchProps & WithStyles;

class AlreadyAuthenticated extends Component<Props> {
  render() {
    const { classes } = this.props;
    return (
      <div className={classes.container}>
        <div className={classes.successContainer}>
          <CircleLoader status={Status.SUCCESSFUL} />
          <span className={classes.messageContainer}>
            <b>{this.props.username}</b><br/>
            you are authenticated
          </span>
        </div>
        <div>Close this tab or logout</div>
        <div className={classes.logoutButtonContainer}>
          <Button
            onClick={this.props.onLogoutClicked}
            variant="contained"
            color="primary">
            Logout
          </Button>
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(AlreadyAuthenticated);