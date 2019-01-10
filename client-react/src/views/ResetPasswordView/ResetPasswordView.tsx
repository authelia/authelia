import React, { Component } from "react";
import { TextField, Button, WithStyles, withStyles } from "@material-ui/core";
import { RouterProps } from "react-router";

import styles from '../../assets/jss/views/ResetPasswordView/ResetPasswordView';

interface Props extends RouterProps, WithStyles {};

class ResetPasswordView extends Component<Props> {
  render() {
    const { classes } = this.props;
    return (
      <div>
        <div>Enter your new password</div>
        <div className={classes.form}>
          <TextField
            className={classes.field}
            variant="outlined"
            id="password1"
            label="New password">
          </TextField>
          <TextField
            className={classes.field}
            variant="outlined"
            id="password2"
            label="Confirm password">
          </TextField>
          <Button
              onClick={() => this.props.history.push('/')}
              variant="contained"
              color="primary"
              className={classes.button}>
              Next
            </Button>
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(ResetPasswordView);