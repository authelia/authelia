import React, { Component } from "react";

import TextField from '@material-ui/core/TextField';

import BottomNavigation from '@material-ui/core/BottomNavigation';
import BottomNavigationAction from '@material-ui/core/BottomNavigationAction';
import RestoreIcon from '@material-ui/core/Icon';
import FavoriteIcon from '@material-ui/core/Icon';
import Button from '@material-ui/core/Button';

import FormTemplate from '../../templates/form-template';

import styles from './second-factor.module.css';
import pendrive from '../../pendrive.png'

interface State {
  mode: number;
}

export class SecondFactor extends Component<any, State> {
  constructor(props: any) {
    super(props);
    this.state = {
      mode: 0
    }
  }

  onMenuChanged(event: any, value: number) {
    this.setState({mode: value});
  }

  renderInner() {
    const registerDevice = (
      <div className={styles.register}>
        <div>Register a new device</div>
        <div className={styles.buttons}>
          <Button variant="contained" color="primary">
            Security key
          </Button>
          <Button variant="contained" color="primary">
            One-time password
          </Button>
        </div>
      </div>
    )

    const authenticate = (
      <div className={styles.authenticate}>
        <div className={styles.u2f}>
          Touch your security key
          <div>
            <img src={pendrive} alt="usb key"/>
          </div>
        </div>
        <table style={{width: '60%', margin: '2em auto'}}>
          <tbody>
            <tr>
              <td style={{width: '40%'}}><hr/></td>
              <td style={{width: '20%'}}>or</td>
              <td style={{width: '40%'}}><hr/></td>
            </tr>
          </tbody>
        </table>
        <div className={styles.totp}>
          Provide a one-time password
          <div className={styles.totpField}>
            <TextField
              id="otp"
              variant="outlined"
              label="Password">
            </TextField>
          </div>
        </div>
      </div>
    )

    if (this.state.mode == 0) {
      return authenticate;
    }
    else if (this.state.mode == 1) {
      return registerDevice;
    }
  }

  render() {
    return (
      <FormTemplate title="2-Factor">
        <div className={styles.main}>
          {this.renderInner()}
        </div>
        <BottomNavigation
            value={this.state.mode}
            onChange={this.onMenuChanged.bind(this)}
            showLabels
            className={styles.menu}
          >
            <BottomNavigationAction label="Authenticate" icon={<RestoreIcon />} />
            <BottomNavigationAction label="Register" icon={<FavoriteIcon />} />
          </BottomNavigation>
      </FormTemplate>
    )
  }
}