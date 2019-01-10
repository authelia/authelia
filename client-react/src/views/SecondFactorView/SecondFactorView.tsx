import React, { Component } from 'react';

import { WithStyles, withStyles, Button, TextField } from '@material-ui/core';

import styles from '../../assets/jss/views/SecondFactorView/SecondFactorView';
import securityKeyImage from '../../assets/images/security-key-hand.png';

type Mode = 'u2f' | 'totp';

interface Props extends WithStyles {};

interface State {
  mode: Mode;
}

class SecondFactorView extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      mode: 'u2f',
    }
  }

  private toggleMode = () => {
    if (this.state.mode === 'u2f') {
      this.setState({mode: 'totp'});
    } else if (this.state.mode === 'totp') {
      this.setState({mode: 'u2f'});
    }
  }

  private renderU2f() {
    const { classes } = this.props;
    return (
      <div>
        <div className={classes.imageContainer}>
          <img src={securityKeyImage} alt='security key' className={classes.image}/>
        </div>
        <div>Insert your security key into a USB port and touch the gold disk.</div>
      </div>
    )
  }

  private renderTotp() {
    const { classes } = this.props;
    return (
      <div>
        <div>Provide a one-time password.</div>
        <TextField
          className={classes.totpField}
          name="password"
          id="password"
          variant="outlined"
          label="Password">
        </TextField>
        <Button
          className={classes.totpButton}
          variant="contained"
          color="primary">
          OK
        </Button>
      </div>
    )
  }

  private renderMode() {
    if (this.state.mode === 'u2f') {
      return this.renderU2f();
    } else if (this.state.mode === 'totp') {
      return this.renderTotp();
    }
  }

  render() {
    const { classes } = this.props;
    return (
      <div className={classes.container}>
        <div className={classes.body}>
          {this.renderMode()}
        </div>
        <hr />
        <div className={classes.footer}>
          <a 
            className={classes.otherMethod}
            href="#"
            onClick={this.toggleMode}>
            {this.state.mode === 'u2f' ? 'Use one-time password' : 'Use security key'}
          </a>
          <a className={classes.registerDevice} href="/security-key-registration">Register device</a>
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(SecondFactorView);