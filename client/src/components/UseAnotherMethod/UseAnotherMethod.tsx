import React, { Component } from 'react';
import styles from '../../assets/scss/components/SecondFactorForm/SecondFactorForm.module.scss';
import Method2FA from '../../types/Method2FA';
import { Button } from '@material/react-button';
import classnames from 'classnames';

export interface OwnProps {}

export interface StateProps {
  availableMethods: Method2FA[] | null;
  isSecurityKeySupported: boolean;
}

export interface DispatchProps {
  onOneTimePasswordMethodClicked: () => void;
  onSecurityKeyMethodClicked: () => void;
  onDuoPushMethodClicked: () => void;
}

export type Props = OwnProps & StateProps & DispatchProps;

interface MethodDescription {
  name: string;
  onClicked: () => void;
  key: Method2FA;
}

class UseAnotherMethod extends Component<Props> {
  render() {
    const methods: MethodDescription[] = [
      {
        name: "One-Time Password",
        onClicked: this.props.onOneTimePasswordMethodClicked,
        key: "totp"
      },
      {
        name: "Security Key (U2F)",
        onClicked: this.props.onSecurityKeyMethodClicked,
        key: "u2f"
      },
      {
        name: "Duo Push Notification",
        onClicked: this.props.onDuoPushMethodClicked,
        key: "duo_push"
      }
    ];

    const methodsComponents = methods
      // Filter out security key if not supported by browser.
      .filter(m => m.key !== "u2f" || (m.key === "u2f" && this.props.isSecurityKeySupported))
      // Filter out the methods that are not supported by the server.
      .filter(m => this.props.availableMethods && this.props.availableMethods.includes(m.key))
      // Create the buttons
      .map(m => <Button raised onClick={m.onClicked} key={m.key}>{m.name}</Button>);

    return (
      <div className={classnames('use-another-method-view')}>
        <div>Choose a method</div>
        <div className={styles.buttonsContainer}>
          {methodsComponents}
        </div>
      </div>
    )
  }
}

export default UseAnotherMethod;