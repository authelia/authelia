import React, { Component } from "react";
import classnames from 'classnames';

import styles from '../../assets/scss/components/Notification/Notification.module.scss';

interface Props {
  className?: string;
  show: boolean;
}

class Notification extends Component<Props> {
  render() {
    return (this.props.show)
      ? (<div className={classnames(styles.container, this.props.className, 'notification')}>
          {this.props.children}
        </div>)
      : null;
  }
}

export default Notification;