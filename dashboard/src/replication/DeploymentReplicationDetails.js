import ReactTimeout from 'react-timeout';
import React, { Component } from 'react';
import api, { IsUnauthorized } from '../api/api.js';
import Loading from '../util/Loading.js';
import styled from 'react-emotion';
import { Header, Loader, Segment } from 'semantic-ui-react';
import { withAuth } from '../auth/Auth.js';
import { Field, FieldContent as FC, FieldLabel as FL } from '../style/style';

const LoaderBox = styled('span')`
  float: right;
  width: 0;
  padding-right: 1em;
  margin-right: 1em;
  margin-top: 1em;
  max-width: 0;
  display: inline-block;
`;

const EndpointView = ({title, deploymentName, masterEndpoint, authKeyfileSecretName, authUserSecretName, tlsCACert, tlsCACertSecretName}) => (
  <Segment>
    <Header>{title}</Header>
    <Field>
      <FL>Deployment</FL>
      <FC>{deploymentName || "-"}</FC>
    </Field>
    <Field>
      <FL>Master endpoint</FL>
      <FC>{masterEndpoint || "-"}</FC>
    </Field>
    <Field>
      <FL>TLS CA Certificate</FL>
      <FC><code>{tlsCACert}</code></FC>
    </Field>
    <Header sub>Secret names</Header>
    <Field>
      <FL>Authentication keyfile</FL>
      <FC><code>{authKeyfileSecretName || "-"}</code></FC>
    </Field>
    <Field>
      <FL>Authentication user</FL>
      <FC><code>{authUserSecretName || "-"}</code></FC>
    </Field>
    <Field>
      <FL>TLS CA Certificate</FL>
      <FC><code>{tlsCACertSecretName || "-"}</code></FC>
    </Field>
  </Segment>
);

const DetailsView = ({replication, loading}) => (
  <div>
    <LoaderBox><Loader size="mini" active={loading} inline/></LoaderBox>
    <EndpointView
      title="Source"
      deploymentName={replication.source.deployment_name}
      masterEndpoint={replication.source.master_endpoint}
      authKeyfileSecretName={replication.source.auth_keyfile_secret_name}
      authUserSecretName={replication.source.auth_user_secret_name}
      tlsCACert={replication.source.tls_ca_cert}
      tlsCACertSecretName={replication.source.tls_ca_cert_secret_name}
    />
    <EndpointView
      title="Destination"
      deploymentName={replication.destination.deployment_name}
      masterEndpoint={replication.destination.master_endpoint}
      authKeyfileSecretName={replication.destination.auth_keyfile_secret_name}
      authUserSecretName={replication.destination.auth_user_secret_name}
      tlsCACert={replication.destination.tls_ca_cert}
      tlsCACertSecretName={replication.destination.tls_ca_cert_secret_name}
    />
  </div>
);

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
    return (<DetailsView replication={dr} loading={this.state.loading}/>);
  }
}

export default ReactTimeout(withAuth(DeploymentReplicationDetails));
