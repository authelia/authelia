import React, { Component } from "react";

import logo from '../logo.svg';
import styles from "./form-template.module.css"

interface Props {
  title: string;
}

export default class FormTemplate extends Component<Props> {
  render() {
    const children = this.props.children;
    return (
      <div className={styles.mainContent}>
        <div className={styles.header}>
          <h1>{this.props.title}</h1>
        </div>
        <div className={styles.frame}>
          <div className={styles.innerFrame}>
            {children}
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