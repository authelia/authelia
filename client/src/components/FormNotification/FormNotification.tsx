import React, { Component } from "react";
import { withStyles, WithStyles, Collapse } from "@material-ui/core";

import styles from '../../assets/jss/components/FormNotification/FormNotification';

interface Props extends WithStyles {
  show: boolean;
}

class FormNotification extends Component<Props> {
  render() {
    const { classes } = this.props;
    return (
      <Collapse in={this.props.show}>
        <div className={classes.messageOuter}>
          <div className={classes.messageInner}>
            <div className={classes.messageContainer}>
              {this.props.children}
            </div>
          </div>
        </div>
      </Collapse>
    )
  }
}

export default withStyles(styles)(FormNotification);