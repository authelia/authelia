import React, { Component } from "react";
import AlreadyAuthenticated from "../../containers/components/AlreadyAuthenticated/AlreadyAuthenticated";
import FirstFactorForm from "../../containers/components/FirstFactorForm/FirstFactorForm";
import SecondFactorForm from "../../containers/components/SecondFactorForm/SecondFactorForm";
import RemoteState from "./RemoteState";
import { RouterProps, Redirect } from "react-router";
import queryString from 'query-string';

export enum Stage {
  FIRST_FACTOR,
  SECOND_FACTOR,
  ALREADY_AUTHENTICATED,
}

export interface StateProps {
  stage: Stage;
  remoteState: RemoteState | null;
  redirectionUrl: string | null;
}

export interface DispatchProps {
  onInit: (redirectionUrl?: string) => void;
}

export type Props = StateProps & DispatchProps & RouterProps;

class AuthenticationView extends Component<Props> {
  componentDidMount() {
    if (this.props.history.location) {
      const params = queryString.parse(this.props.history.location.search);
      if ('rd' in params) {
        this.props.onInit(params['rd'] as string);
      }
    }
    this.props.onInit();
  }

  render() {
    if (!this.props.remoteState) return null;

    if (this.props.stage === Stage.SECOND_FACTOR) {
      return <SecondFactorForm
        username={this.props.remoteState.username}
        redirection={this.props.redirectionUrl} />;
    } else if (this.props.stage === Stage.ALREADY_AUTHENTICATED) {
      return <AlreadyAuthenticated
        username={this.props.remoteState.username}/>;
    }
    return <FirstFactorForm />;
  }
}

export default AuthenticationView;