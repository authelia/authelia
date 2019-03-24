import React, { Component } from 'react';
import styles from '../../assets/scss/components/SecondFactorForm/SecondFactorForm.module.scss';
import Method2FA from '../../types/Method2FA';
import SecondFactorTOTP from '../../containers/components/SecondFactorTOTP/SecondFactorTOTP';
import SecondFactorU2F from '../../containers/components/SecondFactorU2F/SecondFactorU2F';
import classnames from 'classnames';
import SecondFactorDuoPush from '../../containers/components/SecondFactorDuoPush/SecondFactorDuoPush';
import UseAnotherMethod from '../../containers/components/UseAnotherMethod/UseAnotherMethod';

export interface OwnProps {
  username: string;
  redirectionUrl: string | null;
}

export interface StateProps {
  method: Method2FA | null;
  useAnotherMethod: boolean;
}

export interface DispatchProps {
  onInit: () => void;
  onLogoutClicked: () => void;
  onUseAnotherMethodClicked: () => void;
}

export type Props = OwnProps & StateProps & DispatchProps;

class SecondFactorForm extends Component<Props> {
  componentDidMount() {
    this.props.onInit();
  }

  private renderMethod() {
    let method: Method2FA = (this.props.method) ? this.props.method : 'totp'
    let methodComponent, title: string;
    if (method == 'u2f') {
      title = "Security Key";
      methodComponent = (<SecondFactorU2F redirectionUrl={this.props.redirectionUrl}></SecondFactorU2F>);
    } else if (method == "duo_push") {
      title = "Duo Push Notification";
      methodComponent = (<SecondFactorDuoPush redirectionUrl={this.props.redirectionUrl}></SecondFactorDuoPush>);
    } else {
      title = "One-Time Password"
      methodComponent = (<SecondFactorTOTP redirectionUrl={this.props.redirectionUrl}></SecondFactorTOTP>);
    }

    return (
      <div className={classnames('second-factor-step')} key={method + '-method'}>
        <div className={styles.methodName}>{title}</div>
        {methodComponent}
      </div>
    );
  }

  private renderUseAnotherMethodLink() {
    return (
      <div className={styles.anotherMethodLink}>
        <a onClick={this.props.onUseAnotherMethodClicked}>
          Use another method
        </a>
      </div>
    );
  }

  render() {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.hello}>Hello <b>{this.props.username}</b></div>
          <div className={styles.logout}>
            <a onClick={this.props.onLogoutClicked}>Logout</a>
          </div>
        </div>
        <div className={styles.body}>
          {(this.props.useAnotherMethod) ? <UseAnotherMethod/> : this.renderMethod()}
        </div>
        {(this.props.useAnotherMethod) ? null : this.renderUseAnotherMethodLink()}
      </div>
    )
  }
}

export default SecondFactorForm;