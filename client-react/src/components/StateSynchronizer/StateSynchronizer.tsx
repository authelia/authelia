import React, { Component } from "react";
import RemoteState from "../../reducers/Portal/RemoteState";
import { WithState } from "./WithState";

export type OnLoaded = (state: RemoteState) => void;
export type OnError = (err: Error) => void;

export interface Props extends WithState {
  fetch: (onloaded: OnLoaded, onerror: OnError) => void;
  onLoaded: OnLoaded;
  onError?: OnError;
}

class StateSynchronizer extends Component<Props> {
  componentWillMount() {
    this.props.fetch(
      (state) => this.props.onLoaded(state),
      (err: Error) => {
        if (this.props.onError) this.props.onError(err);
      });
  }

  render() {
    return null;
  }
}

export default StateSynchronizer;