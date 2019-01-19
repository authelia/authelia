import React, { Component } from "react";

import { Route, Switch, Redirect, RouterProps, RouteProps } from "react-router";

import { routes } from '../../routes/routes';
import { AUTHELIA_GITHUB_URL } from "../../constants";
import { WithStyles, withStyles } from "@material-ui/core";

import styles from '../../assets/jss/layouts/PortalLayout/PortalLayout';

interface Props extends RouterProps, RouteProps, WithStyles {}

class PortalLayout extends Component<Props> {
  private renderTitle() {
    if (!this.props.location) return;

    for (let i in routes) {
      const route = routes[i];
      if (route.path && route.path.indexOf(this.props.location.pathname) > -1) {
        return route.title.toUpperCase();
      }
    }
    return;
  }


  render() {
    const { classes } = this.props;
    return (
      <div className={classes.mainContent}>
        <div className={classes.frame}>
          <div className={classes.innerFrame}>
            <div className={classes.title}>
              {this.renderTitle()}
            </div>
            <div className={classes.content}>
              <Switch>
                {routes.map((r, key) => {
                  return <Route path={r.path} component={r.component} exact={true} key={key} />
                })}
                <Redirect to='/' />
              </Switch>
            </div>
          </div>
        </div>
        <div className={classes.footer}>
          <div>Powered by <a href={AUTHELIA_GITHUB_URL}>Authelia</a></div>
        </div>
      </div>
    )
  }
}

export default withStyles(styles)(PortalLayout);