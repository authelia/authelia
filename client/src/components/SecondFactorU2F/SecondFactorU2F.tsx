import React from 'react';

import classnames from 'classnames';
import CircleLoader, { Status } from '../../components/CircleLoader/CircleLoader';
import styles from '../../assets/scss/components/SecondFactorU2F/SecondFactorU2F.module.scss';

export interface OwnProps {
  redirectionUrl: string | null;
}

export interface StateProps {
  securityKeyVerified: boolean;
  securityKeyError: string | null;
}

export interface DispatchProps {
  onInit: () => void;
  onRegisterSecurityKeyClicked: () => void;
}

export type Props = StateProps & DispatchProps;

interface State {}

export default class SecondFactorU2F extends React.Component<Props, State> {
  componentWillMount() {
    this.props.onInit();
  }

  render() {
    let u2fStatus = Status.LOADING;
    if (this.props.securityKeyVerified) {
      u2fStatus = Status.SUCCESSFUL;
    } else if (this.props.securityKeyError) {
      u2fStatus = Status.FAILURE;
    }
    return (
      <div className={classnames('security-key-view')}>
        <div>Insert your security key into a USB port and touch the gold disk.</div>
        <div className={styles.imageContainer}>
          <CircleLoader status={u2fStatus}></CircleLoader>
        </div>
        <div className={styles.registerDeviceContainer}>
          <a className={classnames(styles.registerDevice, 'register-u2f')}
            onClick={this.props.onRegisterSecurityKeyClicked}>
            Register new device
          </a>
        </div>
      </div>
    )
  }
}