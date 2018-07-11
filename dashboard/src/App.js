import { Container, Segment, Menu, Message } from 'semantic-ui-react';
import React, { Component } from 'react';
import ReactTimeout from 'react-timeout';

import { withAuth } from './auth/Auth';
import api, { isUnauthorized } from './api/api';
import DeploymentOperator from './deployment/DeploymentOperator';
import DeploymentReplicationOperator from './replication/DeploymentReplicationOperator';
import Loading from './util/Loading';
import NoOperator from './NoOperator';
import StorageOperator from './storage/StorageOperator';

const PodInfoView = ({pod, namespace}) => (
  <Segment basic>
    <Message>
      <Message.Header>Kube-ArangoDB</Message.Header>
      <p>Running in Pod <b>{pod}</b> in namespace <b>{namespace}</b>.</p>
    </Message>
  </Segment>
);

const OperatorsView = ({error, deployment, deploymentReplication, storage, pod, namespace, otherOperators}) => {
  let commonMenuItems = otherOperators.map((item) => <Menu.Item><a href={item.url}>{operatorType2Name(item.type)}</a></Menu.Item>);
  if (commonMenuItems.length > 0) {
    commonMenuItems = (<Menu.Item>
      <Menu.Header>Other operators</Menu.Header>
      <Menu.Menu>{commonMenuItems}</Menu.Menu>
    </Menu.Item>);
  }
  let Operator = NoOperator;
  if (deployment)
    Operator = DeploymentOperator;
  else if (deploymentReplication) 
    Operator = DeploymentReplicationOperator;
  else if (storage)
    Operator = StorageOperator;
  return (
    <Operator
      podInfoView={<PodInfoView pod={pod} namespace={namespace} />}
      commonMenuItems={commonMenuItems}
      error={error}
    />
  );
}

const operatorType2Name = (oType) => {
  switch (oType) {
    case "deployment":
      return "Deployments";
    case "deployment_replication":
      return "Deployment replications";
    case "storage":
      return "Storage";
    default:
      return "";
  }
};

const LoadingView = () => (
  <Container>
    <Loading/>
  </Container>
);

class App extends Component {
  state = {
    operators: undefined,
    error: undefined
  };

  componentDidMount() {
    this.reloadOperators();
  }

  reloadOperators = async() => {
    try {
      const operators = await api.get('/api/operators');
      this.setState({
        operators,
        error: undefined
      });
    } catch (e) {
      this.setState({
        error: e.message
      });
      if (isUnauthorized(e)) {
        this.props.doLogout();
      }
    }
    this.props.setTimeout(this.reloadOperators, 10000);
  }

  render() {
    if (this.state.operators) {
      return <OperatorsView
        error={this.state.error}
        deployment={this.state.operators.deployment}
        deploymentReplication={this.state.operators.deployment_replication}
        storage={this.state.operators.storage}
        otherOperators={this.state.operators.other || []}
        pod={this.state.operators.pod}
        namespace={this.state.operators.namespace}
        />;
    }
    return (<LoadingView/>);
  }
}

export default ReactTimeout(withAuth(App));
