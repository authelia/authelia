import React, { Component } from "react";

import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import logo from '../logo.svg';
import styles from "./index.module.css"

interface State {
  rememberMe: boolean;
}

export class FirstFactor extends Component<any, State> {
  constructor(props: any) {
    super(props)
    this.state = {
      rememberMe: false
    }
  }

  toggleRememberMe = () => {
    this.setState({
      rememberMe: !(this.state.rememberMe)
    })
  }

  render() {
    return (
      <div className={styles.mainContent}>
        <div className={styles.header}>
          <h1>Sign in</h1>
        </div>
        <div className={styles.frame}>
          <div className={styles.innerFrame}>
            <div className={styles.fields}>
              <div className={styles.field}>
                <TextField
                  className={styles.input}
                  id="username"
                  label="Username">
                </TextField>
              </div>
              <div className={styles.field}>
                <TextField
                  className={styles.input}
                  id="password"
                  label="Password"
                  type="password">
                </TextField>
              </div>
            </div>
            <div className={styles.controlArea}>
              <div className={styles.controls}>
                <div className={styles.rememberMe}>
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={this.state.rememberMe}
                        onChange={this.toggleRememberMe}
                        color="primary"
                      />
                    }
                    label="Remember me"
                  />
                </div>
                <div className={styles.resetPassword}>
                  <a href="/">Forgot password?</a>
                </div>
              </div>
              <div className={styles.buttons}>
                <Button
                  variant="contained"
                  color="primary"
                  className={styles.button}>
                  Login
                </Button>
              </div>
            </div>
          </div>
        </div>
        <div className={styles.footer}>
          <img src={logo} alt="logo"></img>
          <div>Powered by <a href="#">Authelia</a></div>
        </div>
      </div>
    )
  }
}