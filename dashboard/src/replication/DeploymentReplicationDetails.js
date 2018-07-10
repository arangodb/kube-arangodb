import ReactTimeout from 'react-timeout';
import React, { Component } from 'react';
import api, { IsUnauthorized } from '../api/api.js';
import Loading from '../util/Loading.js';
import styled from 'react-emotion';
import { Loader } from 'semantic-ui-react';
import { withAuth } from '../auth/Auth.js';

const LoaderBox = styled('span')`
  float: right;
  width: 0;
  padding-right: 1em;
  margin-right: 1em;
  margin-top: 1em;
  max-width: 0;
  display: inline-block;
`;

class DeploymentReplicationDetails extends Component {
  state = {
    loading: true,
    error: undefined
  };

  componentDidMount() {
    this.reloadDeploymentReplications();
  }

  reloadDeploymentReplications = async() => {
    try {
      this.setState({
        loading: true
      });
      const result = await api.get(`/api/deployment-replication/${this.props.name}`);
      this.setState({
        replication: result,
        loading: false,
        error: undefined
      });
    } catch (e) {
      this.setState({
        loading: false,
        error: e.message
      });
      if (IsUnauthorized(e)) {
        this.props.doLogout();
        return;
      }
    }
    this.props.setTimeout(this.reloadDeploymentReplications, 5000);
  }

  render() {
    const dr = this.state.replication;
    if (!dr) {
      return (<Loading/>);
    }
    return (
      <div>
        <LoaderBox><Loader size="mini" active={this.state.loading} inline/></LoaderBox>
        <div>TODO</div>
      </div>
      );
  }
}

export default ReactTimeout(withAuth(DeploymentReplicationDetails));
