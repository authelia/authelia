import React, { Component } from 'react';
import styles from '../../assets/scss/components/SecondFactorForm/SecondFactorForm.module.scss';
import Method2FA from '../../types/Method2FA';
import SecondFactorTOTP from '../../containers/components/SecondFactorTOTP/SecondFactorTOTP';
import SecondFactorU2F from '../../containers/components/SecondFactorU2F/SecondFactorU2F';
import { Button } from '@material/react-button';
import classnames from 'classnames';

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
  onOneTimePasswordMethodClicked: () => void;
  onSecurityKeyMethodClicked: () => void;
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

  private renderUseAnotherMethod() {
    return (
      <div className={classnames('use-another-method-view')}>
        <div>Choose a method</div>
        <div className={styles.buttonsContainer}>
          <Button raised onClick={this.props.onOneTimePasswordMethodClicked}>One-Time Password</Button>
          <Button raised onClick={this.props.onSecurityKeyMethodClicked}>Security Key (U2F)</Button>
        </div>
      </div>
    );
  }

  private renderUseAnotherMethodLink() {
    return (
      <div className={styles.anotherMethodLink}>
        <a href="#" onClick={this.props.onUseAnotherMethodClicked}>
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
            <a onClick={this.props.onLogoutClicked} href="#">Logout</a>
          </div>
        </div>
        <div className={styles.body}>
          {(this.props.useAnotherMethod) ? this.renderUseAnotherMethod() : this.renderMethod()}
        </div>
        {(this.props.useAnotherMethod) ? null : this.renderUseAnotherMethodLink()}
      </div>
    )
  }
}

export default SecondFactorForm;