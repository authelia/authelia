import React, { Component } from "react";
import { TextField, WithStyles, withStyles, Button } from "@material-ui/core";

import styles from '../../assets/jss/views/ForgotPasswordView/ForgotPasswordView';
import { RouterProps } from "react-router";

interface Props extends WithStyles, RouterProps {}

class ForgotPasswordView extends Component<Props> {
  render() {
    const { classes } = this.props;
    return (
      <div>
        <div>What's you e-mail address?</div>
        <div className={classes.form}>
          <TextField
            className={classes.field}
            variant="outlined"
            id="email"
            label="E-mail">
          </TextField>
          <Button
              onClick={() => this.props.history.push('/confirmation-sent')}
              variant="contained"
              color="primary"
              className={classes.button}>
              Next
            </Button>
        </div>
      </div>
    );
  }
}

export default withStyles(styles)(ForgotPasswordView);