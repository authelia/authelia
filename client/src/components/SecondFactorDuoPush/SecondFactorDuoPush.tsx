import React from 'react';

import classnames from 'classnames';
import CircleLoader, { Status } from '../../components/CircleLoader/CircleLoader';
import styles from '../../assets/scss/components/SecondFactorDuoPush/SecondFactorDuoPush.module.scss';
import { Button } from '@material/react-button';

export interface OwnProps {
  redirectionUrl: string | null;
}

export interface StateProps {
  duoPushVerified: boolean | null;
  duoPushError: string | null;
}

export interface DispatchProps {
  onInit: () => void;
  onRetryClicked: () => void;
}

export type Props = OwnProps & StateProps & DispatchProps;

export default class SecondFactorDuoPush extends React.Component<Props> {
  componentWillMount() {
    this.props.onInit();
  }

  render() {
    let u2fStatus = Status.LOADING;
    if (this.props.duoPushVerified === true) {
      u2fStatus = Status.SUCCESSFUL;
    } else if (this.props.duoPushError) {
      u2fStatus = Status.FAILURE;
    }
    return (
      <div className={classnames('duo-push-view')}>
        <div>You will soon receive a push notification on your phone.</div>
        <div className={styles.imageContainer}>
          <CircleLoader status={u2fStatus}></CircleLoader>
        </div>
        {(u2fStatus == Status.FAILURE)
          ? <div className={styles.retryContainer}>
              <Button raised onClick={this.props.onRetryClicked}>Retry</Button>
            </div>
          : null}
      </div>
    )
  }
}