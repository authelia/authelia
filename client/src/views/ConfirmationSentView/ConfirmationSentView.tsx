import React, { Component } from "react";
import classnames from 'classnames';

import mail from '../../assets/images/mail.png';

import styles from '../../assets/scss/views/ConfirmationSentView/ConfirmationSentView.module.scss';
import Button from "@material/react-button";
import { RouterProps } from "react-router";

interface Props extends RouterProps {}

class ConfirmationSentView extends Component<Props> {
  render() {
    return (
      <div className={styles.main}>
        <div className={classnames(styles.image, styles.left)}>
          <img src={mail} alt="mail" />
        </div>
        <div className={styles.right}>
          Please check your e-mails and follow the instructions to confirm the operation.
          <div className={styles.buttonContainer}>
              <Button
                onClick={() => this.props.history.goBack()}
                className={styles.button}
                raised={true}
                color="primary">
                Back
              </Button>
            </div>
        </div>
      </div>
    )
  }
}

export default ConfirmationSentView;