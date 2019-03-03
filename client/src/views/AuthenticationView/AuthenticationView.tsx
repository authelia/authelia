import React, { Component } from "react";
import AlreadyAuthenticated from "../../containers/components/AlreadyAuthenticated/AlreadyAuthenticated";
import FirstFactorForm from "../../containers/components/FirstFactorForm/FirstFactorForm";
import SecondFactorForm from "../../containers/components/SecondFactorForm/SecondFactorForm";
import RemoteState from "./RemoteState";
import { RouterProps, RouteProps } from "react-router";

export enum Stage {
  FIRST_FACTOR,
  SECOND_FACTOR,
  ALREADY_AUTHENTICATED,
}

export interface OwnProps extends RouteProps {}

export interface StateProps {
  stage: Stage;
  remoteState: RemoteState | null;
  redirectionUrl: string | null;
}

export interface DispatchProps {
  onInit: () => void;
}

export type Props = StateProps & DispatchProps & RouterProps;

class AuthenticationView extends Component<Props> {
  componentWillMount() {
    this.props.onInit();
  }

  render() {
    if (!this.props.remoteState) return null;

    if (this.props.stage === Stage.SECOND_FACTOR) {
      return <SecondFactorForm
        username={this.props.remoteState.username}
        redirectionUrl={this.props.redirectionUrl} />;
    } else if (this.props.stage === Stage.ALREADY_AUTHENTICATED) {
      return <AlreadyAuthenticated
        username={this.props.remoteState.username}
        redirectionUrl={this.props.redirectionUrl} />;
    }
    return <FirstFactorForm redirectionUrl={this.props.redirectionUrl} />;
  }
}

export default AuthenticationView;