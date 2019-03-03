import React, { Component } from "react";
import classnames from 'classnames';

import styles from '../../assets/scss/components/CircleLoader/CircleLoader.module.scss';

export enum Status {
  LOADING,
  SUCCESSFUL,
  FAILURE,
}

export interface Props {
  status: Status;
}

class CircleLoader extends Component<Props> {
  render() {
    const containerClasses = [styles.circleLoader];
    const checkmarkClasses = [styles.checkmark, styles.draw];
    const crossClasses = [styles.cross, styles.draw];

    if (this.props.status === Status.SUCCESSFUL) {
      containerClasses.push(styles.loadComplete);
      containerClasses.push(styles.success);
      checkmarkClasses.push(styles.show);
    }
    else if (this.props.status === Status.FAILURE) {
      containerClasses.push(styles.loadComplete);
      containerClasses.push(styles.failure);
      crossClasses.push(styles.show);
    }

    const key = 'container-' + this.props.status;
    
    return (
      <div className={classnames(containerClasses)} key={key}>
        {<div className={classnames(checkmarkClasses)} />}
        {<div className={classnames(crossClasses)} />}
      </div>
    )
  }
}

export default CircleLoader;