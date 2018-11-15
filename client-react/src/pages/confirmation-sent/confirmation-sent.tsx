import React, { Component } from "react";

import FormTemplate from '../../templates/form-template';

import mail from '../../mail.png';

import styles from './confirmation-sent.module.css';

export default class ConfirmationSent extends Component {
  render() {
    return (
      <FormTemplate title="Confirmation e-mail">
        <div className={styles.main}>
          <p>An e-mail has been sent to your address.</p>
          <div className={styles.image}>
            <img src={mail} alt="mail" />
          </div>
          <p>Please click on the link provided in the e-mail to confirm the operation.</p>
        </div>
      </FormTemplate>
    )
  }
}